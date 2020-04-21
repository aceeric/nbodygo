/*
This package provides a basic wrapper around the body collection for the simulation. This project
is a port from Java - and the Java version used a ConcurrentLinkedQueue - which is a concurrent
non-blocking collection with great performance.

The n-body simulation spends 99.999% of its time doing two things: 1) iterating the list of bodies,
and, 2) calculating force. Each body has to interrogate each other body each cycle to compute
gravitational force on itself. So for 2500 bodies in the sim, that's 2,500 X 2,500 iterations = 6,250,000
iterations per frame. If we want to achieve 60 frames per second for a smooth animation, that's
375,000,000 iterations over the body collection every second. The force calculation is what it is: it's
about 30 floating point operations and it probably can't be optimized. So that leaves one place to look
for performance: iterating the body collection.

The body collection is iterated by multiple goroutines concurrently. Each goroutine handles force calculation
for one body at a time so that it can exclusively update that body's cumulative force. So the iteration
needs to support concurrent read. It is also possible to add bodies - either from the gRPC server that
is part of the simulation, or when a body fragments it adds fragments into the body collection.

But 99.999% of the time spent in the body collection is read-only concurrent iteration. I experimented with
several concurrent collections - provided as part of Go, and also from various GitHub repos. But at the end
of the day - nothing came close in performance to plain vanilla array iteration. So this package
provides a collection semantic for the body array that heavily favors read-only concurrent iteration, but
also supports concurrent adds, as well as deletes when a body is set not to exist.

It was not a goal of this project to write a collection class but I was force to by virtue of being
unable to find a non-blocking concurrent collection package with performance equivalent to Java's
ConcurrentLinkedQueue. (I'll keep searching.)
*/
package collection
