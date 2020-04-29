package instrumentation

import "github.com/prometheus/client_golang/prometheus"

//
// None of these are intended to be general purpose - they only support the functionality needed by this project
//
type Counter interface {
	Inc()
}

type CounterVec interface {
	With(labels prometheus.Labels) Counter
}

type Gauge interface {
	Set(val float64)
}

type GaugeVec interface {
	With(labels prometheus.Labels) Gauge
}
