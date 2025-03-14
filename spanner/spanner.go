package spanner

import (
	"context"
	"net/http"

	"github.com/datatrails/go-datatrails-common/logger"
)

// when returning an interface from a function that is injected into a package the interface must
// be defined in only one location and this package achieves that.

// Spanner is an interface to the underlying opentracing span interfaces.
type Spanner interface {
	CarrierFromContext(context.Context, logger.Logger) (map[string]string, logger.Logger, error)
	InjectHTTPHeaders(logger.Logger, http.Header) error
	Logger() logger.Logger
	SetTag(string, any)
	Close()
}

type StartSpanFromContextFunc func(context.Context, logger.Logger, string, ...map[string]any) (Spanner, context.Context)
