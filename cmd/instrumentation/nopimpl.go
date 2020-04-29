package instrumentation

import "github.com/prometheus/client_golang/prometheus"

//
// Contains the NOP implementations
//

func (n *nopCounter) Inc() {}

func (v *nopCounterVec) With(prometheus.Labels) Counter {
	return &nopCounterInstance
}

func (n *nopGauge) Set(float64) {}

func (v *nopGaugeVec) With(prometheus.Labels) Gauge {
	return &nopGaugeInstance
}

