package tracing

import (
	"context"
	"fmt"
	"net/http"

	"github.com/datatrails/go-datatrails-common/logger"
	"github.com/datatrails/go-datatrails-common/spanner"
	opentracing "github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
)

// spanning does the actual work of managing a span
type spanning struct {
	log  logger.Logger
	span Span
}

func (sp *spanning) Close() {
	if sp.log != nil {
		sp.log.Close()
		sp.log = nil
	}
	if sp.span != nil {
		sp.span.Finish()
		sp.span = nil
	}
}

// CarrierFromContext extracts the standard b3 fields from the span context if any exist.
//
// This is used to add these values to the applicationProperties of a servicebus message.
func (sp *spanning) CarrierFromContext(ctx context.Context, log logger.Logger) (map[string]string, logger.Logger, error) {
	carrier := opentracing.TextMapCarrier{}
	err := opentracing.GlobalTracer().Inject(sp.span.Context(), opentracing.TextMap, carrier)
	if err != nil {
		log.Infof("Unable to inject span context: %v", err)
		return nil, log, fmt.Errorf("Unable to inject span context: %v", err)
	}
	log.Debugf("CarrierFromContext: carrier %v", carrier)
	traceID, found := carrier[TraceID]
	if found && traceID != "" {
		log = log.WithIndex(TraceID, traceID)
	} else {
		log.Debugf("CarrierFromContext: %s not found", TraceID)
	}
	log.Debugf("CarrierFromContext: carrier map %v", carrier)
	return carrier, log, nil
}

// InjectHTTPHeaders
func (sp *spanning) InjectHTTPHeaders(log logger.Logger, header http.Header) error {
	err := opentracing.GlobalTracer().Inject(
		sp.span.Context(),
		opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(header),
	)
	if err != nil {
		log.Infof("Unable to inject span context: %v", err)
		return fmt.Errorf("Unable to inject span context: %v", err)
	}
	return nil
}

func (sp *spanning) Logger() logger.Logger {
	return sp.log
}

func (sp *spanning) SetTag(key string, value any) {
	sp.span.SetTag(key, value)
}

// StartSpanFromContext interrogates the context for the presence of a span and returns an opaque handler which
// has a Close() method suitable for a defer command.
//
// attrs is variadic but should only have one or two members. The first member denotes tags that
// label the span. The second member is map of acceptable x-b3 and related fields that must be
// inserted into the span carrier. (only used in azbus/sender).
//
// Most calls to StartSpanFromContext will not use the variadic arguments.
func StartSpanFromContext(
	ctx context.Context,
	log logger.Logger,
	label string,
	attrs ...map[string]any,
) (spanner.Spanner, context.Context) {

	var tags map[string]any
	var carrierAttrs map[string]any
	var span Span
	if len(attrs) > 0 {
		tags = attrs[0]
	}
	if len(attrs) > 1 {
		carrierAttrs = attrs[1]
	}
	if carrierAttrs != nil {
		// The attributes map is passed into the open tracing constructor which
		// extracts any bits it is interested in to use to setup the spans etc.
		// It will ignore anything it doesn't care about. So the filtering of the map
		// is done for us and we don't need to pre-filter it.
		// this creates new root span.
		var opts = []opentracing.StartSpanOption{}
		carrier := opentracing.TextMapCarrier{}

		for k, v := range carrierAttrs {
			// Tracing properties will be strings
			value, ok := v.(string)
			if ok {
				log.Debugf("StartSpanFromContext: carrier: %s, %s", k, value)
				carrier.Set(k, value)
			}
		}
		traceID, found := carrier[TraceID]
		if found && traceID != "" {
			log = log.WithIndex(TraceID, traceID)
		}
		spanCtx, err := opentracing.GlobalTracer().Extract(opentracing.TextMap, carrier)
		if err != nil {
			log.Infof("StartSpanFromContext: Unable to extract span context: %v", err)
		} else {
			opts = append(opts, opentracing.ChildOf(spanCtx))
		}
		span = opentracing.StartSpan(label, opts...)
		log.Debugf("StartSpanFromContext: carrier %v", carrier)
		ctx = opentracing.ContextWithSpan(ctx, span)
	} else {
		// if no attributes map is passed in the it is assumed that we want to start a new
		// child span from the context.
		span, ctx = opentracing.StartSpanFromContext(ctx, label)
		log = LogFromContext(ctx, log)
	}

	// Add tags if required.
	if tags != nil {
		logFields := make([]otlog.Field, 0, len(tags))
		for k, v := range tags {
			// Tracing fields will be strings
			value, ok := v.(string)
			if ok {
				log.Debugf("StartSpanFromContext: logField: %s, %s", k, value)
				logFields = append(logFields, otlog.String(k, value))
			}
		}
		log.Debugf("StartSpanFromContext: logFields: %v", logFields)
		span.LogFields(logFields...)
	}
	log.Infof("Span %s with started from context", label)
	return &spanning{span: span, log: log}, ctx
}
