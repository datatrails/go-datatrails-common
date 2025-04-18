// Package tracing is responsible for forwarding and translating span headers for internal requests
package tracing

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"os"
	"strings"
	"time"

	"github.com/datatrails/go-datatrails-common/logger"

	otnethttp "github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
	opentracinglog "github.com/opentracing/opentracing-go/log"

	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	zipkin "github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
)

const (
	requestID         = "x-request-id"
	otSpanContext     = "x-ot-span-context"
	prefixTracerState = "x-b3-"
	TraceID           = prefixTracerState + "traceid"
	spanID            = prefixTracerState + "spanid"
	parentSpanID      = prefixTracerState + "parentspanid"
	sampled           = prefixTracerState + "sampled"
	flags             = prefixTracerState + "flags"
)

var otHeaders = []string{
	requestID,
	otSpanContext,
	prefixTracerState,
	TraceID,
	spanID,
	parentSpanID,
	sampled,
	flags,
}

func SetSpanField(ctx context.Context, key string, value string) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		span.LogFields(
			opentracinglog.String(key, value),
		)
	}
}

func SetSpanHTTPHeader(span opentracing.Span, log logger.Logger, r *http.Request) {
	// Transmit the span's TraceContext as HTTP headers on our request
	if span != nil {
		err := opentracing.GlobalTracer().Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(r.Header),
		)
		if err != nil {
			log.Infof("Tracer.Inject %v", err)
		}
	}
}

func HTTPMiddleware(h http.Handler) http.Handler {
	return otnethttp.Middleware(
		opentracing.GlobalTracer(),
		h,
		otnethttp.OperationNameFunc(func(r *http.Request) string {
			return "HTTP " + r.Method + ":" + r.URL.EscapedPath() + " >"
		}),
	)
}

// HeaderMatcher ensures that open tracing headers x-b3-* are forwarded to output requests
func HeaderMatcher(key string) (string, bool) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	for _, tracingKey := range otHeaders {
		if strings.ToLower(key) == tracingKey {
			return key, true
		}
	}
	return "", false
}

func trimPodName(p string) string {
	a := strings.Split(p, "-")
	i := len(a)

	// We want the pod name without the trailing instance ID components
	// There can be either two ID components (length 10 or 11 and 5) or
	// just one (length 5)

	if len(a[(i-1)]) == 5 && (len(a[(i-2)]) == 10 || len(a[(i-2)]) == 11) {
		// this has two instnace ID components so strip them
		return strings.Join(a[:i-2], "-")
	}
	if i > 1 {
		// otherwise just strip one
		return strings.Join(a[:i-1], "-")
	}
	return p
}

// NewTracer and the following methods are only called once to create the global tracer
// in startup/run.go
func NewTracer(log logger.Logger, portName string) io.Closer {
	instanceName, _, _ := strings.Cut(getOrFatal("POD_NAME"), " ")
	nameSpace := getOrFatal("POD_NAMESPACE")
	containerName := getOrFatal("CONTAINER_NAME")
	podName := strings.Join([]string{trimPodName(instanceName), nameSpace, containerName}, ".")
	listenStr := fmt.Sprintf("localhost:%s", getOrFatal(portName))
	return NewFromEnv(log, strings.TrimSpace(podName), listenStr, "ZIPKIN_ENDPOINT", "DISABLE_ZIPKIN")
}

// NewFromEnv initialises tracing and returns a closer if tracing is
// configured.  If the necessary configuration is not available it is Fatal
// unless disableVar is set and is truthy (strconf.ParseBool -> true). If
// tracing is disabled returns nil
func NewFromEnv(log logger.Logger, service string, host string, endpointVar, disableVar string) io.Closer {
	ze, ok := os.LookupEnv(endpointVar)
	if !ok {
		if disabled := getTruthyOrFatal(disableVar); !disabled {
			log.Panicf(
				"'%s' has not been provided and is not disabled by '%s'",
				endpointVar, disableVar)
		}
		log.Infof("zipkin disabled by '%s'", disableVar)
		return nil
	}
	// zipkin conf is available, disable it if disableVar is truthy

	if disabled := getTruthyOrFatal(disableVar); disabled {
		log.Infof("'%s' set, zipkin disabled", disableVar)
		return nil
	}
	return New(log, service, host, ze)
}

// New initialises tracing
// uses zipkin client tracer
func New(log logger.Logger, service string, host string, zipkinEndpoint string) io.Closer {
	// create our local service endpoint
	localEndpoint, err := zipkin.NewEndpoint(service, host)
	if err != nil {
		log.Panicf("unable to create zipkin local endpoint service '%s' - host '%s': %v", service, host, err)
	}

	// set up a span reporter
	zipkinLogger := newZipkinLogger()
	reporter := zipkinhttp.NewReporter(zipkinEndpoint, zipkinhttp.Logger(zipkinLogger))

	// TODO: One day this should probably be configurable in helm for each service
	// For now capture 1 in every 5 traces
	rate := 0.2

	// This sampler is only used when a service creates new traces (which is rare, only if
	// not recieving messages or presenting callable endpoints, e.g. a cron like service)
	sampler, err := zipkin.NewBoundarySampler(rate, time.Now().UnixNano())
	if err != nil {
		log.Panicf("unable to create zipkin sampler: rate %f: %v", rate, err)
	}

	// initialise the tracer
	nativeTracer, err := zipkin.NewTracer(
		reporter,
		zipkin.WithLocalEndpoint(localEndpoint),
		zipkin.WithSharedSpans(false),
		zipkin.WithSampler(sampler),
	)
	if err != nil {
		log.Panicf("unable to create zipkin tracer: %v", err)
	}

	// use zipkin-go-opentracing to wrap our tracer
	tracer := zipkinot.Wrap(nativeTracer)
	opentracing.SetGlobalTracer(tracer)

	return reporter
}
