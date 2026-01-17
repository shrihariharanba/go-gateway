package openTelemetry

import (
	"context"
	"net/http"

	"github.com/shrihariharanba/go-gateway/internal/telemetry/providers"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

type OTelProvider struct {
	cfg    providers.Config
	tp     *sdktrace.TracerProvider
	closer func(context.Context) error
}

func New(cfg providers.Config) providers.TelemetryProvider {
	return &OTelProvider{cfg: cfg}
}

func (o *OTelProvider) Init(ctx context.Context) error {
	// Example: Use OTLP over HTTP (common for collector setups)
	exp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(o.cfg.Endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(o.cfg.ServiceName),
		),
	)
	if err != nil {
		return err
	}

	o.tp = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(o.tp)
	o.closer = o.tp.Shutdown

	return nil
}

func (o *OTelProvider) Middleware(next http.Handler) http.Handler {
	return otelhttp.NewHandler(next, "gateway-request")
}

func (o *OTelProvider) Handler() http.Handler {
	// No HTTP metric handler for OTEL directly, usually collector handles it
	return nil
}

func (o *OTelProvider) Name() string { return "opentelemetry" }
