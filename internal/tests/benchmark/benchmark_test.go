package benchmark

import (
	"testing"

	"github.com/uber-go/tally"
	"go.uber.org/cff"
	"go.uber.org/zap"
)

// BenchmarkBaseline is a flow that has two concurrent tasks that do almost nothing, that is designed to try to measure
// the overhead incurred by cff.Flow
func BenchmarkBaseline(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Baseline()
	}
}

// BenchmarkBaselineNative is a re-implementation of the Simple flow that makes the most optimal use of Go synchronization primitives
// while still running the two tasks in parallel. It should serve as a baseline as comparison to the Simple function.
func BenchmarkBaselineNative(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BaselineNative()

	}
}

// BenchmarkPredicateCombined emulates the state of cff.Flow before
// https://code.uberinternal.com/D5495165. This serves as the baseline
// implementation in comparison to BenchmarkPredicateSplit.
func BenchmarkPredicateCombined(b *testing.B) {
	for i := 0; i < b.N; i++ {
		PredicateCombined()
	}
}

// BenchmarkPredicateSplit is a flow that deploys a cff.Task with a
// cff.Predicate option.
func BenchmarkPredicateSplit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		PredicateSplit()
	}
}

type metricsTestFn func(*zap.Logger, tally.Scope) float64

// BenchmarkMetrics is the same flow as Simple but with instrumentation added.
func BenchmarkMetrics(b *testing.B) {
	logger := zap.NewNop()
	scope := tally.NoopScope
	builder := cff.TallyEmitter(scope)

	metricsMemoized := func(logger *zap.Logger, scope tally.Scope) float64 {
		return FlowMetricsMemoized1000(logger, builder)
	}
	metricsFailedMemoized := func(logger *zap.Logger, scope tally.Scope) float64 {
		return Metrics1000FailedMemoized(logger, builder)
	}

	metricsCases := []struct {
		name string
		fn   metricsTestFn
	}{
		{
			"Metrics", Metrics,
		},
		{
			"MetricsWithSharedEmitter", MetricsWithSharedEmitter,
		},
		{
			"Metrics100", Metrics100,
		},
		{
			"Metrics500", Metrics500,
		},
		{
			"Metrics1000", Metrics1000,
		},
		{
			"Metrics1000Failed", Metrics1000Failed,
		},
		{
			"MetricsMemoized1000", metricsMemoized,
		},
		{
			"Metrics1000FailedMemoized", metricsFailedMemoized,
		},
	}

	for _, metricsCase := range metricsCases {
		metricsCaseClosure := metricsCase
		b.Run(metricsCase.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				metricsCaseClosure.fn(logger, scope)
			}
		})
	}
}
