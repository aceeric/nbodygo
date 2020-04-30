package instrumentation

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"time"
)

type nopCounter struct{}
type nopCounterVec struct{}
type nopGauge struct{}
type nopGaugeVec struct{}

var nopCounterInstance = nopCounter{}
var nopCounterVecInstance = nopCounterVec{}
var nopGaugeInstance = nopGauge{}
var nopGaugeVecInstance = nopGaugeVec{}

var server *http.Server

const address = ":12345"

//
// Returns a NOP counter or a Prometheus counter based on the init function
//
func NewCounter(opts prometheus.CounterOpts) Counter {
	if isPrometheus {
		return promauto.NewCounter(opts)
	} else {
		return &nopCounterInstance
	}
}

//
// Returns a NOP CounterVec or a wrapped Prometheus CounterVec based on the init function
//
func NewCounterVec(opts prometheus.CounterOpts, labelNames []string) CounterVec {
	if isPrometheus {
		return &promCounterVec{
			promauto.NewCounterVec(opts, labelNames),
		}
	} else {
		return &nopCounterVecInstance
	}
}

//
// Returns a NOP Gauge or a Prometheus Gauge based on the init function
//
func NewGauge(opts prometheus.GaugeOpts) Gauge {
	if isPrometheus {
		return promauto.NewGauge(opts)
	} else {
		return &nopGaugeInstance
	}
}

//
// Returns a NOP GaugeVec or a wrapped Prometheus GaugeVec based on the init function
//
func NewGaugeVec(opts prometheus.GaugeOpts, labelNames []string) GaugeVec {
	if isPrometheus {
		return &promGaugeVec{
			promauto.NewGaugeVec(opts, labelNames),
		}
	} else {
		return &nopGaugeVecInstance
	}
}

//
// Starts the Prometheus HTTP server
//
func Start() {
	if isPrometheus {
		server = &http.Server{Addr: address, Handler: promhttp.Handler()}
		go func() {
			if err := server.ListenAndServe(); err != nil {
				log.Fatalf("Instrumentation was unable to start Prometheus HTTP server. The error is: %v\n", err)
			}
		}()
	}
}

//
// Stops the Prometheus HTTP server
//
func Stop() {
	if server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Print("[ERROR] Unable to stop Prometheus HTTP")
		}
		server = nil
	}
}
