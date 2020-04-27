/*
In the Java application, the simulation instantiated the gRPC server and passed
it a reference to itself so the gRPC server could call back into the simulation:

+-----+                     +-------------+
| sim | -- create(this) --> | gRPC server |
+-----+                     +-------------+

Then:
                  +-------------+
User  -- gRPC --> | gRPC server |
                  +-------------+
+-----+                 |
| sim |<-- call --------+
+-----+

The simulation and the gRPC server were in different Java packages. In the port to Go, they are
still different (Go) packages. This is a natural paradigm in every other programming language
I've ever used. But in Go, this results in a circular import. There seem to be a couple of
recommended ways to resolve this. I elected to define a struct of callback functions in this, an
intermediary package, and initialize the callbacks in the sim package. Then, the callback struct is
provided to the gRPC server package. The gRPC interface method handlers then invoke the appropriate
callbacks in the struct to call into the sim package.

As a result, the package dependencies are no longer circular (to Go):

+-----+                   +------------+
| sim | -- depends on --> | grpcserver | -- depends on --+
+-----+                   +------------+                 |
   |                      +------------+                 |
   +-- depends on ------> | grpcsimcb  | <---------------+
                          | (this pkg) |
                          +------------+

So Go is happy. This seems fairly tortuous but Go's definition of "circular" dependencies
mandates a solution from a small set of recommended approaches.
*/
package grpcsimcb
