package azbus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/google/uuid"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/datatrails/go-datatrails-common/tracing"
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

	log                   Logger
	mtx                   sync.Mutex
	sender                *azservicebus.Sender
	maxMessageSizeInBytes int64
}

// NewSender creates a new client
func NewSender(log Logger, cfg SenderConfig) *Sender {

	s := &Sender{
		Cfg:      cfg,
		azClient: NewAZClient(cfg.ConnectionString),
	}
	s.log = log.WithIndex("sender", s.String())
	return s
}

func (s *Sender) String() string {
	return s.Cfg.TopicOrQueueName
}

func (s *Sender) Close(ctx context.Context) {

	var err error
	if s != nil && s.sender != nil {
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

// Send submits a message to the queue. Ignores cancellation.
func (s *Sender) Send(ctx context.Context, message *OutMessage) error {

	var err error

	// boots & braces
	if s.sender == nil {
		err = s.Open()
		if err != nil {
			return err
		}
	}

	// Without this fix eventsourcepoller and similar services repeatedly context cancel and repeatedly
	// restart.
	ctx = context.WithoutCancel(ctx)

	span, ctx := tracing.StartSpanFromContext(ctx, "Sender.Send")
	defer span.Finish()

	// Get the logging context after we create the span as that may have created a new
	// trace and stashed the traceid in the metadata.
	log := tracing.LogFromContext(ctx, s.log)
	defer log.Close()

	// We set and log a message ID so we can trace the message through the bus
	id := uuid.New().String()
	message.MessageID = &id

	span.LogFields(
		otlog.String("sender", s.String()),
		otlog.String("message id", id),
	)

	size := int64(len(message.Body))
	log.Debugf("%s: Msg id %s Sized %d limit %d", s, id, size, s.maxMessageSizeInBytes)
	if size > s.maxMessageSizeInBytes {
		log.Debugf("Msg Sized %d > limit %d :%v", size, s.maxMessageSizeInBytes, ErrMessageOversized)
		return fmt.Errorf("%s: Msg Sized %d > limit %d :%w", s, size, s.maxMessageSizeInBytes, ErrMessageOversized)
	}
	now := time.Now()

	s.updateSendingMesssageForSpan(ctx, message, span)

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

// BatchAddMessage calls Addmessage on batch and adds span tracing
func (s *Sender) BatchAddMessage(ctx context.Context, span tracing.Span, batch *OutMessageBatch, m *OutMessage, options *azservicebus.AddMessageOptions) error {
	s.updateSendingMesssageForSpan(ctx, m, span)
	return batch.AddMessage(m, options)
}

// SendBatch submits a message batch to the broker. Ignores cancellation.
func (s *Sender) SendBatch(ctx context.Context, batch *OutMessageBatch) error {

	// Without this fix eventsourcepoller and similar services repeatedly context cancel and repeatedly
	// restart.
	ctx = context.WithoutCancel(ctx)

	var err error

	now := time.Now()

	span, ctx := tracing.StartSpanFromContext(ctx, "Sender.SendBatch")
	defer span.Finish()
	span.LogFields(
		otlog.String("sender", s.String()),
	)

	// Get the logging context after we create the span as that may have created a new
	// trace and stashed the traceid in the metadata.
	log := tracing.LogFromContext(ctx, s.log)
	defer log.Close()

	// boots & braces
	if s.sender == nil {
		err = s.Open()
		if err != nil {
			return err
		}
	}

	// Note: sizing must be dealt with as the batch is created and accumulated.
	err = s.sender.SendMessageBatch(ctx, batch, nil)
	if err != nil {
		azerr := fmt.Errorf("SendMessageBatch failed in %s: %w", time.Since(now), NewAzbusError(err))
		log.Infof("%s", azerr)
		return azerr
	}
	return nil
}
