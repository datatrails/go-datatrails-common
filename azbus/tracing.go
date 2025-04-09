package azbus

import (
	"context"

	"github.com/datatrails/go-datatrails-common/tracing"
	opentracing "github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
)

func CreateReceivedMessageTracingContext(ctx context.Context, log Logger, label string, message *ReceivedMessage) (context.Context, opentracing.Span) {
	// We don't have the tracing span info on the context yet, that is what this function will add
	// we we log using the reciever logger
	log.Debugf("ContextFromReceivedMessage(): ApplicationProperties %v", message.ApplicationProperties)

	var opts = []opentracing.StartSpanOption{}
	carrier := opentracing.TextMapCarrier{}
	// This just gets all the message Application Properties into a string map. That map is then passed into the
	// open tracing constructor which extracts any bits it is interested in to use to setup the spans etc.
	// It will ignore anything it doesn't care about. So the filtering of the map is done for us and
	// we don't need to pre-filter it.
	for k, v := range message.ApplicationProperties {
		// Tracing properties will be strings
		value, ok := v.(string)
		if ok {
			carrier.Set(k, value)
		}
	}
	spanCtx, err := opentracing.GlobalTracer().Extract(opentracing.TextMap, carrier)
	if err != nil {
		log.Infof("CreateReceivedMessageWithTracingContext(): Unable to extract span context: %v", err)
	} else {
		opts = append(opts, opentracing.ChildOf(spanCtx))
	}
	span := opentracing.StartSpan(label, opts...)
	ctx = opentracing.ContextWithSpan(ctx, span)
	return ctx, span
}

func (r *Receiver) handleReceivedMessageWithTracingContext(ctx context.Context, message *ReceivedMessage, label string, handler Handler) (Disposition, context.Context, error) {
	ctx, span := CreateReceivedMessageTracingContext(ctx, r.log, label, message)
	defer span.Finish()
	span.LogFields(
		otlog.String("receiver", r.String()),
	)
	return handler.Handle(ctx, message)
}

func (s *Sender) updateSendingMesssageForSpan(ctx context.Context, message *OutMessage, span tracing.Span) {
	log := tracing.LogFromContext(ctx, s.log)
	defer log.Close()

	carrier := opentracing.TextMapCarrier{}
	err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.TextMap, carrier)
	if err != nil {
		log.Infof("updateSendingMesssageForSpan(): Unable to inject span context: %v", err)
		return
	}
	for k, v := range carrier {
		OutMessageSetProperty(message, k, v)
	}
	log.Debugf("updateSendingMesssageForSpan(): ApplicationProperties %v", OutMessageProperties(message))
}
