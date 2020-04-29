# N-Body Go

This project is a port - in progress - of https://github.com/aceeric/nbodyjava from **Java** into the **Go** language.

The application is an implementation of the "n-body" physics problem in which a space is filled with a number of bodies (in this case spheres) that each have mass, radius, and velocity. The n-body problem requires that each body accumulate its force from each other body continuously in a perpetual loop. (Called the *brute-force* approach.)

This is a relatively straightforward problem and there are well-documented solutions. In fact, my project incorporates code modeled from http://physics.princeton.edu/~fpretori/Nbody/code.htm to actually perform the force calculation. The interesting thing about the project is to design the simulation to support as many bodies as possible in real time. This is where the challenges are, and what makes it interesting.

Since every body accumulates force from every other body, the number of computations increases significantly with the number of bodies because on each cycle through the bodies in the simulation, n * n-1 force calculations occur. For example:

| Bodies                       | 100        | 500         | 1,000         | 2,000         | 5,000          |
| :---------------------------- | ----------: | -----------: | -------------: | -------------: | --------------: |
| Force calculations per cycle | 9,999      | 249,999     | 999,999       | 3,999,999     | 24,999,999     |
| Flops per force calculation  | 30         | 30          | 30            | 30            | 30             |
| Flops per cycle              | 299,970    | 7,499,970   | 29,999,970    | 119,999,970   | 749,999,970    |
| Frames per second            | 50         | 50          | 50            | 50            | 50             |
| Flops per second             | 14,998,500 | 374,998,500 | 1,499,998,500 | 5,999,998,500 | 37,499,998,500 |

Looking at the column for 1K bodies: it requires 999 thousand force calculations for one cycle through all the bodies (1000 * 999). Each force calculation is about 30 floating point operations. So that's 29.9 million flops per cycle. The objective is to render the simulation in real time which requires about 50 frames per second resulting in 1.5 billion floating point operations per second to calculate the gravitational attraction for 1K bodies. Going to 2K bodies increases flops per second to 6 billion. And so on.

And that doesn't include collision resolution, which requires substantially more floating point operations per collision.

My goal for this project was to build a Golang app that was 100% feature-compatible with the Java version. The Java version is fully documented - you can take a look at the GitHub repo referenced at the top of this README. The Java version used a game engine called JMonkey Engine (https://jmonkeyengine.org/) to render the simulation. I looked around at the options for a Go-based game engine and - short story - settled on something called G3N (http://g3n.rocks/) as the Go rendering engine.

it's safe to say that Go isn't a platform widely used for game development. So it was nice to find a Go-based game engine that supported the set of functionality that I used from JMonkey which - granted - is a limited set of the features offered by both engines.

The Go project design is generally consistent with the Java design: there is a *server* Go app that runs the simulation, opens up a G3N window on the desktop, and displays the sim results. The reason it is referred to as a server is that the executable contains a gRPC server that listens for messages from a client app. (See https://grpc.io/ for more info about gRPC.)

The *client* Go app allows you to add bodies into a running simulation, and change sim characteristics.

As in the Java version, I have only tested this under Ubuntu 18.04. I may get around to testing under Windows at some point.

This repo includes the G3N Go code. I made a couple of minor changes to support my needs and it seemed simplest to just include their code - along with the requisite license file. (See internal/g3n/LICENSE).

### Prerequisites

The following steps are required to build and run the application. This README assumes that you've already installed Go. If that's *not* the case, please refer to https://golang.org/doc/install.  This project was developed using Go 1.14. This README also assumes the availability of Gnu Make, which is almost universally available on Linux.

First clone the GitHub repository to your local machine:

```bash
$ git clone https://github.com/aceeric/nbodygo.git
$ cd nbodygo
```

The following prerequisites are documented on https://grpc.io/docs/quickstart/go/:

Install gRPC as a Go module:

```bash
$ export GO111MODULE=on
$ go get google.golang.org/grpc@v1.28.1
```

Install the gRPC protobuf compiler. This is the tool that converts the gRPC interface description language into Go code that is referenced by both the client and the server. The protobuf compiler is referenced by the project Make file. On Ubuntu:

```bash
$ sudo apt install protobuf-compiler
```

If that succeeds, then you can execute `which protoc` to get the location of the installed binary, and add it to your PATH variable.

Install the Go `protoc` plugin and update your PATH as documented on the quick start:

```bash
$ go get github.com/golang/protobuf/protoc-gen-go
$ export PATH="$PATH:$(go env GOPATH)/bin"
```

You have to install some requirements for G3N, as documented on their GitHub repo: https://github.com/g3n/engine:

```bash
$ sudo apt install xorg-dev libgl1-mesa-dev libopenal1 libopenal-dev libvorbis0a libvorbis-dev libvorbisfile3
```

Now, you should be able to build the application (assumes you're in the `nbodygo` directory that you git cloned):

```bash
$ make
```

This should create the client and the server executables in `bin/client` and `bin/server`, respectively:

```bash
$ ls -l bin
total 33116
-rwxr-xr-x 1 eace eace 14495244 Apr 28 15:06 client
-rwxr-xr-x 1 eace eace 19411256 Apr 28 15:06 server
```

Finally, you can run the server with no arguments which runs a default canned sim:

```bash
$ bin/server
```

This should start the simulation window with 1,000 bodies orbiting a sun. You can use the standard keyboard navigation keys to navigate the sim: W A S D, etc. See the N-Body Java GitHub page for details, and all the command-line options and gRPC client options.

### Java vs Go?

TODO

### Simulations

TODO

### gRPC client differences

My intent was to keep the gRPC interfaces identical between the Go and Java versions. But I wound up making a couple small changes to the Go version. One particular difference results from how G3N implements light sources. To support that, the Go gRPC client `add-body` command supports an `intensity` property. Also, because Go is more explicit with type conversion than Java, I wound up changing some data types in the gRPC interface to reduce the number of type conversions in the code. As a result, you can't use the Go gRPC client with the Java server, or vice versa.

### To Do

The following tasks remain to complete the project:

| Task            | Description                                                  |
| :-------------- | ------------------------------------------------------------ |
| Instrumentation | Copy additional/grafana, additional/prometheus, and additional/scripts/start-containers into this repo. Modify the Grafana dashboard to support Go metrics, remove JVM metrics. Update README |
| Simulations     | Copy additional/sims from Java project, and update to use the bin/client go executable with correct syntax |
| README          | Add instructions for running the client and server. (Copy from Java?) |
| Logging         | The Java app used Log4j for logging. Go has built in logging and so that needs to be added in to take the place of the Log4j |
| Tests           | The Java version didn't include unit tests. The Go version should have those added |
| To Do           | There are numerous `todo` comments sprinkled throughout the code to be cleaned up |
| Go docs         | Finalize the go docs                                         |
| Windows?        | Assess effort to support Windows                             |

