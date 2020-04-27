/*
The runner package provides the mechanics of running the n-body computation using a worker pool,
and interfacing with the gRPC service to modify the simulation in a concurrent, thread-safe
way while the sim is running. There are three abstractions.

ComputationRunner

The computation runner submits the n-body computation to the work pool, and then collects the
results and places them in a ResultQueue.

WorkPool

The work pool runs a pool of go routines that receive work on a channel. The goal was to provide the
same functionality that was available in the Java version using a Java thread pool executor.

ResultQueue

The result queue allows the computation runner to outrun the rendering by a fixed amount. In practice this
turns out to be of limited use since the rendering engine enforces a frame rate and the computation runner
can't handle more than about 3K bodies without falling below that rate. May replace this with an A/B
list at some point. The main goal was to provide the rendering engine with a list of renderable objects
while the computation runner had concurrent access to the actual bodies so the computation and the
rendering can run concurrently.
*/
package runner
