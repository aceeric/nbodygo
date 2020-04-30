/*
This package encapsulates how the project uses instrumentation. The project supports Prometheus
instrumentation, and NOP instrumentation. The idea is that you add instrumentation calls into the
project, but they are all NOP calls unless you start the server with the 'NBODYPROM' environment
variable defined. E.g.:

$ NBODYPROM= bin/server &

Defining this variable causes the instrumentation package to start the embedded Prometheus HTTP server, and
causes all instrumentation calls to actually instantiate and register Prometheus metric collectors. If you
don't specify that environment variable, no HTTP server is started, and all the metric calls wind up at
NOP methods that do nothing. So - while you incur the call overhead to run without instrumentation,
you don't occur any additional overhead.

The goal is to be able to leave the instrumentation calls in the code but not incur the associated overhead
unless needed, since most of the time you want the app to run with as much CPU devoted to the sim as possible.
Granted, there probably isn't a use case for this when instrumenting a production service, but for development
or troubleshooting, perhaps it has value.
*/
package instrumentation
