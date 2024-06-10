package main

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// # of queries run, the total
// processing time across all queries, the minimum query time (for a single query), the median
// query time, the average query time, and the maximum query time.

type Stats interface {
	Record(duration time.Duration)
	Print() error
	Stop()
}

type stats struct {
	// TotalQueries int
	// TotalProcessingTime float64
	// MinQueryTime float64
	// MaxQueryTime float64
	// // Median
	histogram metric.Int64Histogram

	ot innerOtel
}

type innerOtel struct {
	resource      *resource.Resource
	reader        *metricsdk.ManualReader
	meterProvider *metricsdk.MeterProvider
	meter         metric.Meter
}

func NewStats() (Stats, error) {
	res, err := resource.Merge(resource.Default(),
		resource.NewSchemaless(
			semconv.ServiceName("tsdb-exercise"),
			semconv.ServiceVersion("0.1.0"),
		))
	if err != nil {
		return nil, err
	}

	reader := metricsdk.NewManualReader()

	meterProvider := metricsdk.NewMeterProvider(
		metricsdk.WithResource(res),
		metricsdk.WithReader(reader),
	)

	otel.SetMeterProvider(meterProvider)

	meter := otel.Meter("tsdb-benchmark")

	histogram, err := meter.Int64Histogram(
		"query.duration",
		metric.WithDescription("The duration of queries execution."),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return nil, err
	}

	return &stats{
		ot: innerOtel{
			resource:      res,
			reader:        reader,
			meterProvider: meterProvider,
			meter:         meter,
		},
		histogram: histogram,
	}, nil
}

func (s *stats) Record(duration time.Duration) {
	s.histogram.Record(context.Background(), duration.Milliseconds())
}

func (s *stats) Print() error {
	data := metricdata.ResourceMetrics{}
	err := s.ot.reader.Collect(context.Background(), &data)
	if err != nil {
		return err
	}

	exporter, err := stdoutmetric.New()
	if err != nil {
		return err
	}
	ctx := context.Background()
	defer exporter.Shutdown(ctx)
	err = exporter.Export(ctx, &data)
	histogram := data.ScopeMetrics[0].Metrics[0]
	fmt.Println(histogram)
	return err
}

func (s *stats) Stop() {
	ctx := context.Background()
	s.ot.meterProvider.Shutdown(ctx)
	s.ot.reader.Shutdown(ctx)
}
