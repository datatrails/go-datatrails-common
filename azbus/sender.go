package azbus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	otrace "github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
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
		log:      log,
		azClient: NewAZClient(cfg.ConnectionString),
	}
	return s
}

func (s *Sender) String() string {
	return s.Cfg.TopicOrQueueName
}

func (s *Sender) GetAZClient() AZClient {
	return s.azClient
}

func (s *Sender) Close(ctx context.Context) {

	var err error
	if s != nil && s.sender != nil {
		s.log.Debugf("Close %s", s)
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

	azadmin := NewAZAdminClient(s.log, s.Cfg.ConnectionString)
	s.maxMessageSizeInBytes, err = azadmin.GetQueueMaxMessageSize(s.Cfg.TopicOrQueueName)
	if err != nil {
		azerr := fmt.Errorf("%s: failed to get sender properties: %w", s, NewAzbusError(err))
		s.log.Infof("%s", azerr)
		return azerr
	}
	s.log.Debugf("%s: Maximum message size is %d bytes", s, s.maxMessageSizeInBytes)

	sender, err := client.NewSender(s.Cfg.TopicOrQueueName, nil)
	if err != nil {
		azerr := fmt.Errorf("%s: failed to open sender: %w", s, NewAzbusError(err))
		s.log.Infof("%s", azerr)
		return azerr
	}

	s.log.Debugf("Open Sender %s", s)
	s.sender = sender
	return nil
}

type OutMessageOption func(*OutMessage)

func WithProperty(key string, value any) OutMessageOption {
	return func(o *OutMessage) {
		o.ApplicationProperties[key] = value
	}
}

func (s *Sender) Send(ctx context.Context, msg []byte, opts ...OutMessageOption) error {
	return s.SendMsg(ctx, NewOutMessage(msg), opts...)
}

// Send submits a message to the queue. Ignores cancellation.
func (s *Sender) SendMsg(ctx context.Context, message OutMessage, opts ...OutMessageOption) error {

	// Without this fix eventsourcepoller and similar services repeatedly context cancel and repeatedly
	// restart.
	ctx = contextWithoutCancel(ctx)

	var err error

	log := s.log.FromContext(ctx)
	defer log.Close()

	span, ctx := otrace.StartSpanFromContext(ctx, "Sender.Send")
	defer span.Finish()
	span.LogFields(
		otlog.String("sender", s.Cfg.TopicOrQueueName),
	)

	err = s.Open()
	if err != nil {
		return err
	}

	size := int64(len(message.Body))
	log.Debugf("%s: Msg Sized %d limit %d", s, size, s.maxMessageSizeInBytes)
	if size > s.maxMessageSizeInBytes {
		log.Debugf("%s: Msg Sized %d > limit %d :%v", s, size, s.maxMessageSizeInBytes, ErrMessageOversized)
		return fmt.Errorf("%s: Msg Sized %d > limit %d :%w", s, size, s.maxMessageSizeInBytes, ErrMessageOversized)
	}
	now := time.Now()
	if message.ApplicationProperties == nil {
		message.ApplicationProperties = make(map[string]any)
	}
	opts = AddCorrelationIDOption(ctx, opts...)
	for _, opt := range opts {
		opt(&message)
	}
	log.Debugf("ApplicationProperties %v", message.ApplicationProperties)
	err = s.sender.SendMessage(ctx, &message, nil)
	if err != nil {
		azerr := fmt.Errorf("%s: Send failed in %s: %w", s, time.Since(now), NewAzbusError(err))
		log.Infof("%s", azerr)
		return azerr
	}
	log.Debugf("%s: Sending message took %s", s, time.Since(now))
	return nil
}
