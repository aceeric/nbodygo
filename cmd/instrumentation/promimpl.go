package instrumentation

import "github.com/prometheus/client_golang/prometheus"

//
// Contains the prometheus implementations
//

type promCounterVec struct {
	vec *prometheus.CounterVec
}

func (cv *promCounterVec) With(labels prometheus.Labels) Counter {
	return cv.vec.With(labels)
}

type promGaugeVec struct {
	vec *prometheus.GaugeVec
}

func (cv *promGaugeVec) With(labels prometheus.Labels) Gauge {
	return cv.vec.With(labels)
}
