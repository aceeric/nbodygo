package instrumentation

import (
	"github.com/prometheus/client_golang/prometheus"
)

//
// locates all the application-wide metrics here so they're not sprinkled throughout the app. (The
// references are sprinkled but the instantiations are all here)
//

var BodyComputations Counter
var NoComputationQueues Counter
var NoRenderQueues Counter
var ComputationCount CounterVec
var ComputationWorkers Gauge
var MaxQueues Gauge
var CurQueues Gauge
var BodyCount GaugeVec

func InitMetrics() {
	BodyComputations = NewCounter(
		prometheus.CounterOpts{
			Name: "nbody_computations",
			Help: "N-Body nested loop iterations",
		},
	)
	NoComputationQueues = NewCounter(
		prometheus.CounterOpts{
			Name: "nbody_no_computation_queues_count",
			Help: "Count of computation runner outrunning rendering engine",
		},
	)
	NoRenderQueues = NewCounter(
		prometheus.CounterOpts{
			Name: "nbody_no_queues_to_render_count",
			Help: "Count of rendering engine outrunning computation runner",
		},
	)
	ComputationCount = NewCounterVec(
		prometheus.CounterOpts{
			Name: "nbody_computation_count",
			Help: "Simulation cycles for runner and renderer",
		},
		[]string{"thread"},
	)
	ComputationWorkers = NewGauge(
		prometheus.GaugeOpts{
			Name: "nbody_computation_thread_gauge",
			Help: "Computation Runner worker pool size",
		},
	)
	BodyCount = NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nbody_body_count_gauge",
			Help: "Number of bodies in the simulation",
		},
		[]string{"thread"},
	)
	MaxQueues = NewGauge(
		prometheus.GaugeOpts{
			Name: "nbody_result_queue_max_size",
			Help: "Max cached computation results",
		},
	)
	CurQueues = NewGauge(
		prometheus.GaugeOpts{
			Name: "nbody_result_queue_size",
			Help: "Current cached computation results",
		},
	)
}
