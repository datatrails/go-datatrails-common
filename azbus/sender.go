package azbus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/datatrails/go-datatrails-common/logger"
	"github.com/datatrails/go-datatrails-common/spanner"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/google/uuid"
)

// SenderConfig configuration for an azure servicebus namespace and queue
type SenderConfig struct {
	ConnectionString string

	// Name is the name of the queue or topic to send to.
	TopicOrQueueName string
}

// Sender to send or receive messages on  a queue or topic
type Sender struct {
	azClient AZClient

	Cfg SenderConfig

	log                   logger.Logger
	mtx                   sync.Mutex
	sender                *azservicebus.Sender
	maxMessageSizeInBytes int64
	spanner               spanner.StartSpanFromContextFunc
}

type SenderOption func(*Sender)

func WithSenderTracing(f spanner.StartSpanFromContextFunc) SenderOption {
	return func(s *Sender) {
		s.spanner = f
	}
}

// NewSender creates a new client
func NewSender(log logger.Logger, cfg SenderConfig, opts ...SenderOption) *Sender {

	s := Sender{
		Cfg:      cfg,
		azClient: NewAZClient(cfg.ConnectionString),
	}
	s.log = log.WithIndex("sender", s.String())
	for _, opt := range opts {
		opt(&s)
	}
	return &s
}

func (s *Sender) String() string {
	return s.Cfg.TopicOrQueueName
}

func (s *Sender) Close(ctx context.Context) {

	var err error
	if s == nil {
		return
	}
	if s.sender != nil {
		s.log.Debugf("Close")
		s.mtx.Lock()
		defer s.mtx.Unlock()
		err = s.sender.Close(ctx)
		if err != nil {
			azerr := fmt.Errorf("%s: Error closing sender: %w", s, NewAzbusError(err))
			s.log.Infof("%s", azerr)
		}
		s.sender = nil // not going to attempt to close again on error
	}
}

func (s *Sender) Open() error {
	var err error

	if s.sender != nil {
		return nil
	}

	client, err := s.azClient.azClient()
	if err != nil {
		return err
	}

	azadmin := newazAdminClient(s.log, s.Cfg.ConnectionString)
	s.maxMessageSizeInBytes, err = azadmin.getQueueMaxMessageSize(s.Cfg.TopicOrQueueName)
	if err != nil {
		azerr := fmt.Errorf("%s: failed to get sender properties: %w", s, NewAzbusError(err))
		s.log.Infof("%s", azerr)
		return azerr
	}
	s.log.Debugf("Maximum message size is %d bytes", s.maxMessageSizeInBytes)

	sender, err := client.NewSender(s.Cfg.TopicOrQueueName, nil)
	if err != nil {
		azerr := fmt.Errorf("%s: failed to open sender: %w", s, NewAzbusError(err))
		s.log.Infof("%s", azerr)
		return azerr
	}

	s.log.Debugf("Open")
	s.sender = sender
	return nil
}

func (*Sender) updateSendingMessage(ctx context.Context, sp spanner.Spanner, log logger.Logger, message *OutMessage) logger.Logger {

	var err error
	var attrs map[string]string

	attrs, log, err = sp.CarrierFromContext(ctx, log)
	if err != nil {
		log.Infof("updateSendingMessage(): failed to get carrier: %v", err)
	}
	for k, v := range attrs {
		OutMessageSetProperty(message, k, v)
	}
	log.Debugf("updateSendingMessage: ApplicationProperties %v", OutMessageProperties(message))
	return log
}

// Send submits a message to the queue. Ignores cancellation.
func (s *Sender) Send(ctx context.Context, message *OutMessage) error {

	log := s.log

	ctx = context.WithoutCancel(ctx)
	var span spanner.Spanner
	var err error

	id := uuid.New().String()
	message.MessageID = &id

	if s.spanner != nil {
		span, ctx = s.spanner(
			ctx,
			log,
			"Sender.Send",
			map[string]any{
				"sender":     s.Cfg.TopicOrQueueName,
				"message.id": id,
			},
		)
		defer span.Close()
	}

	log.Debugf("Span Sender.Send")
	// boots & braces
	if s.sender == nil {
		err = s.Open()
		if err != nil {
			return err
		}
	}

	size := int64(len(message.Body))
	log.Debugf("%s: Msg id %s Sized %d limit %d", s, id, size, s.maxMessageSizeInBytes)
	if size > s.maxMessageSizeInBytes {
		log.Debugf("Msg Sized %d > limit %d :%v", size, s.maxMessageSizeInBytes, ErrMessageOversized)
		return fmt.Errorf("%s: Msg Sized %d > limit %d :%w", s, size, s.maxMessageSizeInBytes, ErrMessageOversized)
	}
	if span != nil {
		log = s.updateSendingMessage(ctx, span, log, message)
	}

	now := time.Now()
	err = s.sender.SendMessage(ctx, message, nil)
	if err != nil {
		azerr := fmt.Errorf("Send message id %s failed in %s: %w", id, time.Since(now), NewAzbusError(err))
		log.Infof("%s", azerr)
		return azerr
	}
	log.Debugf("Sending message id %s took %s", id, time.Since(now))
	return nil
}

func (s *Sender) NewMessageBatch(ctx context.Context) (*OutMessageBatch, error) {
	return s.sender.NewMessageBatch(ctx, nil)
}

// BatchAddMessage calls Addmessage on batch
// Note: this method is a direct pass through and exists only to provide a
// mockable interface for adding messages to a batch.
func (s *Sender) BatchAddMessage(batch *OutMessageBatch, m *OutMessage, options *azservicebus.AddMessageOptions) error {
	return batch.AddMessage(m, options)
}

// SendBatch submits a message batch to the broker. Ignores cancellation.
func (s *Sender) SendBatch(ctx context.Context, batch *OutMessageBatch) error {

	log := s.log

	// Without this fix eventsourcepoller and similar services repeatedly context cancel and repeatedly
	// restart.
	ctx = context.WithoutCancel(ctx)

	var span spanner.Spanner
	var err error

	if s.spanner != nil {
		span, ctx = s.spanner(
			ctx,
			log,
			"Sender.SendBatch",
			map[string]any{
				"sender": s.Cfg.TopicOrQueueName,
			},
		)
		defer span.Close()
		log = span.LogFromContext(ctx, log)
	}
	log.Debugf("Span Sender.SendBatch")

	// boots & braces
	if s.sender == nil {
		err = s.Open()
		if err != nil {
			return err
		}
	}
	// Note: sizing must be dealt with as the batch is created and accumulated.

	// Note: the first message properties (including application properties) are established by the first message in the batch

	now := time.Now()
	err = s.sender.SendMessageBatch(ctx, batch, nil)
	if err != nil {
		azerr := fmt.Errorf("SendMessageBatch failed in %s: %w", time.Since(now), NewAzbusError(err))
		log.Infof("%s", azerr)
		return azerr
	}
	return nil
}
