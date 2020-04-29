package instrumentation

import (
	"os"
)

//
// E.g.:
// $ NBODYPROM= ./bin/server &
//
// ...will enable prometheus instrumentation for the n-body server. (There is no
// instrumentation for the client)
//
const nBodyPrometheusEnvVar = "NBODYPROM"

var isPrometheus = false

func init() {
	_, isPrometheus = os.LookupEnv(nBodyPrometheusEnvVar)
	InitMetrics()
}
