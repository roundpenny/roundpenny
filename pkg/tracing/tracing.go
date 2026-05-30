// Copyright (c) 2026 RoundPenny. All rights reserved.

package tracing

import (
	"context"
	"os"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func InitTracing(serviceName string) (*trace.TracerProvider, error) {
	endpoint := os.Getenv("OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "tempo:4317"
	}

	sampleRate := 0.1
	if v := os.Getenv("TRACING_SAMPLE_RATE"); v != "" {
		if r, err := strconv.ParseFloat(v, 64); err == nil {
			sampleRate = r
		}
	}

	exporter, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(semconv.ServiceName(serviceName)),
	)
	if err != nil {
		return nil, err
	}

	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.TraceIDRatioBased(sampleRate)),
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	return tp, nil
}
