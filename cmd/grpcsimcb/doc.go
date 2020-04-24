/*
In the Java application, the simulation instantiated the gRPC server and passed
it a reference to itself so the gRPC server could call back into the simulation:

+-----+                     +-------------+
| sim | -- create(this) --> | gRPC server |
+-----+                     +-------------+

Then:
                        +-------------+
User  -- change sim --> | gRPC server |--call--+
                        +-------------+        |
+-----+                                        |
| sim |<---------------------------------------+
+-----+

The simulation and the gRPC server were in different Java packages. In the port to Go, they are
still different (Go) packages. This is a natural paradigm in every other programming language
I've ever used. But in Go, this results in a circular import. There seem to be a couple of
recommended ways to resolve this. I elected to define a struct of callback functions, and
initialize the functions in the sim package, and then pass the struct to the gRPC server package.
The gRPC interface method handlers then invoke the appropriate callback in the struct.

This seems fairly tortuous but Go's arbitrary refusal to deal with "circular" dependencies
mandates a solution from a small set of recommended approaches.
*/
package grpcsimcb
