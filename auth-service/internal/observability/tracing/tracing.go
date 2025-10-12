package tracing

import (
	"auth-service/internal/config"
	"auth-service/internal/logging"
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"time"
)

// Init sets up an OpenTelemetry exporter if enabled and reachable; otherwise, it keeps OTEL's no-op.
func Init(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	if !config.Current().Observability.TracingConfig.Enabled {
		logging.Logger.Info().Msg("tracing disabled; using no-op tracer")
		return func(context.Context) error { return nil }, nil
	}

	// Create resource (with service name, environment, etc.).
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			attribute.String("env", config.Current().Env),
		),
	)
	if err != nil {
		logging.Logger.Warn().Err(err).Msg("tracing: resource initialization failed; falling back to no-op")
		return func(context.Context) error { return nil }, nil
	}

	// Create the OTLP exporter directly using the endpoint and gRPC connection
	logging.Logger.Info().Str("endpoint", config.Current().Observability.TracingConfig.Endpoint).Msg("stablishing connection to OTEL")
	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(config.Current().Observability.TracingConfig.Endpoint),
		otlptracegrpc.WithInsecure(), // Or remove this if using secure connection
	)
	if err != nil {
		logging.Logger.Warn().Str("endpoint", config.Current().Observability.TracingConfig.Endpoint).
			Err(err).Msg("tracing: collector not reachable; using no-op")
		return func(context.Context) error { return nil }, nil
	}

	// Initialize a tracer provider with the exporter
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exp), // Uses batch exporter to send traces
	)
	otel.SetTracerProvider(tp)

	logging.Logger.Info().Str("endpoint", config.Current().Observability.TracingConfig.Endpoint).Msg("tracing initialized (OTLP)")

	// Return shutdown function for the tracer provider.
	return tp.Shutdown, nil
}

// StartSpan starts a new trace span.
func StartSpan(ctx context.Context, methodName string) (trace.Span, context.Context) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, methodName)
	span.SetAttributes(attribute.String("method", methodName))
	return span, ctx
}

// EndSpan ends the trace span.
func EndSpan(span trace.Span) {
	span.End()
}

// AddAttributesToSpan adds additional attributes to the span.
func AddAttributesToSpan(span trace.Span, attributes map[string]string) {
	for key, value := range attributes {
		span.SetAttributes(attribute.String(key, value))
	}
}

// TODO: Remove needs to be removed ...
func InitHttp(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	headers := map[string]string{
		"content-type": "application/json",
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracehttp.NewClient(
			otlptracehttp.WithEndpoint("localhost:4318"),
			otlptracehttp.WithHeaders(headers),
			otlptracehttp.WithInsecure(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating new exporter: %w", err)
	}

	tracerprovider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(
			exporter,
			sdktrace.WithMaxExportBatchSize(sdktrace.DefaultMaxExportBatchSize),
			sdktrace.WithBatchTimeout(sdktrace.DefaultScheduleDelay*time.Millisecond),
			sdktrace.WithMaxExportBatchSize(sdktrace.DefaultMaxExportBatchSize),
		),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(serviceName),
			),
		),
	)

	otel.SetTracerProvider(tracerprovider)

	return tracerprovider.Shutdown, nil
}
