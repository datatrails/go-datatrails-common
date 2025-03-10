package tracing

import (
	"context"

	"github.com/datatrails/go-datatrails-common/logger"
	opentracing "github.com/opentracing/opentracing-go"
)

// LogFromContext returns a logger indexed according to the x-b3-traceID key.
func LogFromContext(ctx context.Context, log logger.Logger) logger.Logger {

	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return log
	}
	carrier := opentracing.TextMapCarrier{}
	err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.TextMap, carrier)
	if err != nil {
		log.Infof("LogFromContext: can't inject span: %v", err)
		return log
	}

	traceID, found := carrier[TraceID]
	if !found || traceID == "" {
		log.Debugf("%s not found", TraceID)
		return log
	}
	return log.WithIndex(TraceID, traceID)
}
