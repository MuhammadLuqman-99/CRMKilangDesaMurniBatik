// Package tracer provides distributed tracing utilities using OpenTelemetry.
package tracer

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/kilang-desa-murni/crm/pkg/config"
	"github.com/kilang-desa-murni/crm/pkg/logger"
)

// Tracer wraps the OpenTelemetry tracer.
type Tracer struct {
	provider *sdktrace.TracerProvider
	tracer   trace.Tracer
	config   *config.TracerConfig
	log      *logger.Logger
}

// New creates a new tracer with the given configuration.
func New(cfg *config.TracerConfig, log *logger.Logger) (*Tracer, error) {
	if !cfg.Enabled {
		log.Info().Msg("Tracing is disabled")
		return &Tracer{
			config: cfg,
			log:    log,
			tracer: otel.Tracer(cfg.ServiceName),
		}, nil
	}

	// Create OTLP exporter
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(cfg.Endpoint),
		otlptracehttp.WithInsecure(),
	)

	exporter, err := otlptrace.New(context.Background(), client)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create sampler based on sample rate
	var sampler sdktrace.Sampler
	if cfg.SampleRate >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else if cfg.SampleRate <= 0 {
		sampler = sdktrace.NeverSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(cfg.SampleRate)
	}

	// Create tracer provider
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set global tracer provider
	otel.SetTracerProvider(provider)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	log.Info().
		Str("service", cfg.ServiceName).
		Str("endpoint", cfg.Endpoint).
		Float64("sample_rate", cfg.SampleRate).
		Msg("Tracing initialized")

	return &Tracer{
		provider: provider,
		tracer:   provider.Tracer(cfg.ServiceName),
		config:   cfg,
		log:      log,
	}, nil
}

// Tracer returns the underlying OpenTelemetry tracer.
func (t *Tracer) Tracer() trace.Tracer {
	return t.tracer
}

// Start starts a new span.
func (t *Tracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name, opts...)
}

// StartSpan starts a new span with common options.
func (t *Tracer) StartSpan(ctx context.Context, name string, attributes ...attribute.KeyValue) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name, trace.WithAttributes(attributes...))
}

// SpanFromContext returns the span from the context.
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// AddEvent adds an event to the current span.
func AddEvent(ctx context.Context, name string, attributes ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, trace.WithAttributes(attributes...))
}

// SetAttributes sets attributes on the current span.
func SetAttributes(ctx context.Context, attributes ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attributes...)
}

// RecordError records an error on the current span.
func RecordError(ctx context.Context, err error, attributes ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err, trace.WithAttributes(attributes...))
}

// SetStatus sets the status of the current span.
func SetStatus(ctx context.Context, code trace.SpanKind, description string) {
	span := trace.SpanFromContext(ctx)
	span.SetStatus(trace.SpanKind(code), description)
}

// Close shuts down the tracer provider.
func (t *Tracer) Close(ctx context.Context) error {
	if t.provider == nil {
		return nil
	}

	t.log.Info().Msg("Shutting down tracer")
	return t.provider.Shutdown(ctx)
}

// Common attribute keys
var (
	AttrTenantID    = attribute.Key("tenant.id")
	AttrUserID      = attribute.Key("user.id")
	AttrRequestID   = attribute.Key("request.id")
	AttrHTTPMethod  = attribute.Key("http.method")
	AttrHTTPURL     = attribute.Key("http.url")
	AttrHTTPStatus  = attribute.Key("http.status_code")
	AttrDBStatement = attribute.Key("db.statement")
	AttrDBOperation = attribute.Key("db.operation")
	AttrDBTable     = attribute.Key("db.table")
	AttrEventType   = attribute.Key("event.type")
	AttrEventID     = attribute.Key("event.id")
)

// TenantID creates a tenant ID attribute.
func TenantID(id string) attribute.KeyValue {
	return AttrTenantID.String(id)
}

// UserID creates a user ID attribute.
func UserID(id string) attribute.KeyValue {
	return AttrUserID.String(id)
}

// RequestID creates a request ID attribute.
func RequestID(id string) attribute.KeyValue {
	return AttrRequestID.String(id)
}

// HTTPMethod creates an HTTP method attribute.
func HTTPMethod(method string) attribute.KeyValue {
	return AttrHTTPMethod.String(method)
}

// HTTPURL creates an HTTP URL attribute.
func HTTPURL(url string) attribute.KeyValue {
	return AttrHTTPURL.String(url)
}

// HTTPStatus creates an HTTP status code attribute.
func HTTPStatus(code int) attribute.KeyValue {
	return AttrHTTPStatus.Int(code)
}

// DBStatement creates a database statement attribute.
func DBStatement(stmt string) attribute.KeyValue {
	return AttrDBStatement.String(stmt)
}

// DBOperation creates a database operation attribute.
func DBOperation(op string) attribute.KeyValue {
	return AttrDBOperation.String(op)
}

// DBTable creates a database table attribute.
func DBTable(table string) attribute.KeyValue {
	return AttrDBTable.String(table)
}

// EventType creates an event type attribute.
func EventType(eventType string) attribute.KeyValue {
	return AttrEventType.String(eventType)
}

// EventID creates an event ID attribute.
func EventID(id string) attribute.KeyValue {
	return AttrEventID.String(id)
}
