package tracing

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type otelTracer struct {
	tracer trace.Tracer
	tp     *sdktrace.TracerProvider
	opts   *Options
}

type otelSpan struct {
	span trace.Span
}

// NewOTelTracer 创建一个基于 OpenTelemetry 的追踪器
func NewOTelTracer(opts ...Option) (Tracer, error) {
	options := &Options{
		ServiceName:    "unknown",
		ServiceVersion: "unknown",
		Environment:    "unknown",
		Endpoint:       "localhost:4317",
		Sampler:        1.0,
		Timeout:        5 * time.Second,
		RetryCount:     3,
		RetryDelay:     time.Second,
	}

	for _, opt := range opts {
		opt(options)
	}

	// 创建 OTLP exporter
	ctx := context.Background()
	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(options.Endpoint),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithDialOption(grpc.WithBlock()),
		otlptracegrpc.WithTimeout(options.Timeout),
		otlptracegrpc.WithRetry(otlptracegrpc.RetryConfig{
			Enabled:         true,
			InitialInterval: options.RetryDelay,
			MaxInterval:     options.RetryDelay * 2,
			MaxElapsedTime:  options.Timeout,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %v", err)
	}

	// 创建资源
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(options.ServiceName),
			semconv.ServiceVersion(options.ServiceVersion),
			semconv.DeploymentEnvironment(options.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %v", err)
	}

	// 创建追踪器提供者
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(options.Sampler)),
	)

	// 设置全局追踪器提供者
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer := tp.Tracer(
		options.ServiceName,
		trace.WithInstrumentationVersion(options.ServiceVersion),
	)

	return &otelTracer{
		tracer: tracer,
		tp:     tp,
		opts:   options,
	}, nil
}

func (t *otelTracer) Start(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	options := &SpanOptions{
		StartTime:  time.Now(),
		Kind:       SpanKindInternal,
		Attributes: make([]attribute.KeyValue, 0),
	}

	for _, opt := range opts {
		opt(options)
	}

	spanOpts := []trace.SpanStartOption{
		trace.WithAttributes(options.Attributes...),
		trace.WithTimestamp(options.StartTime),
	}

	switch options.Kind {
	case SpanKindServer:
		spanOpts = append(spanOpts, trace.WithSpanKind(trace.SpanKindServer))
	case SpanKindClient:
		spanOpts = append(spanOpts, trace.WithSpanKind(trace.SpanKindClient))
	case SpanKindProducer:
		spanOpts = append(spanOpts, trace.WithSpanKind(trace.SpanKindProducer))
	case SpanKindConsumer:
		spanOpts = append(spanOpts, trace.WithSpanKind(trace.SpanKindConsumer))
	default:
		spanOpts = append(spanOpts, trace.WithSpanKind(trace.SpanKindInternal))
	}

	ctx, span := t.tracer.Start(ctx, name, spanOpts...)
	return ctx, &otelSpan{span: span}
}

func (t *otelTracer) Extract(ctx context.Context, carrier interface{}) context.Context {
	switch c := carrier.(type) {
	case propagation.TextMapCarrier:
		return otel.GetTextMapPropagator().Extract(ctx, c)
	default:
		return ctx
	}
}

func (t *otelTracer) Inject(ctx context.Context, carrier interface{}) error {
	switch c := carrier.(type) {
	case propagation.TextMapCarrier:
		otel.GetTextMapPropagator().Inject(ctx, c)
		return nil
	default:
		return fmt.Errorf("unsupported carrier type: %T", carrier)
	}
}

func (s *otelSpan) End() {
	s.span.End()
}

func (s *otelSpan) SetName(name string) {
	s.span.SetName(name)
}

func (s *otelSpan) SetStatus(code int, description string) {
	switch code {
	case 0:
		s.span.SetStatus(codes.Ok, "")
	case 1:
		s.span.SetStatus(codes.Error, description)
	default:
		s.span.SetStatus(codes.Unset, "")
	}
}

func (s *otelSpan) SetError(err error) {
	s.span.SetStatus(codes.Error, err.Error())
	s.span.RecordError(err)
}

func (s *otelSpan) SetAttributes(kv ...attribute.KeyValue) {
	s.span.SetAttributes(kv...)
}

func (s *otelSpan) AddEvent(name string, attributes ...attribute.KeyValue) {
	s.span.AddEvent(name, trace.WithAttributes(attributes...))
}

func (s *otelSpan) RecordError(err error, opts ...trace.EventOption) {
	s.span.RecordError(err, opts...)
}

func (s *otelSpan) SpanContext() trace.SpanContext {
	return s.span.SpanContext()
}

func (s *otelSpan) IsRecording() bool {
	return s.span.IsRecording()
}

func (s *otelSpan) TracerProvider() trace.TracerProvider {
	return s.span.TracerProvider()
}
