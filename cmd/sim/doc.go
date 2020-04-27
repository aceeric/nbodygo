/*
The 'sim' package contains the simulation runner and related functionality. It can generate a few
canned simulations, and also generate a simulation from a CSV. Then it runs the simulation, starting
up all the supporting services and goroutines. It waits for the simulation to end, and then shuts
everything down.

The main go file is 'nbodysim.go' which is the sim runner that is invoked by the server main
function.
*/
package sim
