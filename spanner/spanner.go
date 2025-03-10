package spanner

import (
	"context"
	"net/http"

	"github.com/datatrails/go-datatrails-common/logger"
)

// When returning an interface from a function that is injected into a package the interface must
// be defined in only one location and this package achieves that.

// Spanner is an interface to the underlying opentracing span interfaces.
type Spanner interface {
	CarrierFromContext(context.Context, logger.Logger) (map[string]string, logger.Logger, error)
	InjectHTTPHeaders(logger.Logger, http.Header) error
	LogFromContext(context.Context, logger.Logger) logger.Logger
	SetTag(string, any)
	Close()
}

// StartSpanFromContextFunc is injected into any package struct to enable tracing.
type StartSpanFromContextFunc func(context.Context, logger.Logger, string, ...map[string]any) (Spanner, context.Context)
