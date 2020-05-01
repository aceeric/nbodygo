package main

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"nbodygo/cmd/globals"
	"nbodygo/cmd/grpcserver"
	"nbodygo/cmd/nbodygrpc"
	"os"
	"strconv"
	"strings"
	"time"
)

//
// gRPC client state
//
type client struct {
	conn     *grpc.ClientConn
	nbodyCli nbodygrpc.NBodyServiceClient
}

// local server (hard-coded in the server too)
const nbodyServer string = "localhost:50051"

//
// Creates a new gRPC client connection to the gRPC server
//
func newClient() *client {
	var opts = []grpc.DialOption{grpc.WithBlock(), grpc.WithInsecure()}
	conn, err := grpc.Dial(nbodyServer, opts...)
	if err != nil {
		log.Fatalf("Error connecting to gRPC server: %v", err)
	}
	c := client{
		conn:     conn,
		nbodyCli: nbodygrpc.NewNBodyServiceClient(conn),
	}
	return &c
}

//
// Stops the client
//
func (c *client) stopClient() {
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

//
// Sets the number of computation threads on the server
//
func (c *client) setComputationThreads(threadsArg string) {
	threads, err := strconv.ParseInt(threadsArg, 0, 64)
	if err != nil {
		log.Fatalf("Invalid value: %v\n", threadsArg)
	}
	result, _ := c.nbodyCli.SetComputationThreads(context.Background(), &nbodygrpc.ItemCount{ItemCount: threads})
	println(result.Message)
}

//
// Sets the result queue size on the server to the passed value
//
func (c *client) setResultQueueSize(sizeArg string) {
	val, err := strconv.ParseInt(sizeArg, 0, 64)
	if err != nil {
		log.Fatalf("Invalid value: %v\n", sizeArg)
	}
	result, _ := c.nbodyCli.SetResultQueueSize(context.Background(), &nbodygrpc.ItemCount{ItemCount: val})
	println(result.Message)
}

//
// Sets the smoothing factor on the server to the passed value
//
func (c *client) setSmoothing(smoothingArg string) {
	factor, err := strconv.ParseFloat(smoothingArg, 64)
	if err != nil {
		log.Fatalf("Invalid value: %v\n", smoothingArg)
	}
	result, _ := c.nbodyCli.SetSmoothing(context.Background(), &nbodygrpc.Factor{Factor: factor})
	println(result.Message)
}

//
// Sets the coefficient of restitution on the server to the passed value
//
func (c *client) setRestitutionCoefficient(coeffArg string) {
	coeff, err := strconv.ParseFloat(coeffArg, 64)
	if err != nil {
		log.Fatalf("Invalid value: %v\n", coeffArg)
	}
	result, _ := c.nbodyCli.SetRestitutionCoefficient(context.Background(),
		&nbodygrpc.RestitutionCoefficient{RestitutionCoefficient: coeff})
	println(result.Message)
}

//
// Removes the passed number of bodies from the sim on the server. If -1, removes all, including pinned bodies
//
func (c *client) removeBodies(countArg string) {
	count, err := strconv.ParseInt(countArg, 0, 64)
	if err != nil {
		log.Fatalf("Invalid value: %v\n", countArg)
	}
	result, _ := c.nbodyCli.RemoveBodies(context.Background(), &nbodygrpc.ItemCount{ItemCount: count})
	println(result.Message)
}

//
// Gets sim configuration settings from the server and displays them to the console
//
func (c *client) getCurrentConfig() {
	config, err := c.nbodyCli.GetCurrentConfig(context.Background(), &empty.Empty{})
	if err != nil {
		log.Fatalf("get-config error: %v\n", err)
	}
	result := "" +
		"Bodies = %v\n" +
		"Result Queue Size = %v\n" +
		"Computation Threads = %v\n" +
		"Smoothing Factor = %v\n" +
		"Restitution Coefficient = %v\n"
	fmt.Printf(result, config.Bodies, config.ResultQueueSize, config.ComputationThreads, config.SmoothingFactor,
		config.RestitutionCoefficient)
}

//
// Gets a body from the sim on the server and displays its properties to the console
//
// args:
//   whichArg  Specifies which body to get. Either 'id=' or 'name='. E.g.: get-body id=123. Or:
//             get-body name=jupiter
//
func (c *client) getBody(whichArg string) {
	which := strings.Split(whichArg, "=")
	var id int
	var name string
	var err error
	if len(which) != 2 {
		log.Fatalf("get-body needs id=n or name=foo\n")
	}
	if which[0] == "id" {
		id, err = strconv.Atoi(which[1])
		name = ""
	} else if which[0] == "name" {
		name = which[1]
		id = -1
	} else {
		log.Fatalf("get-body needs id=n or name=foo\n")
	}
	if err != nil {
		log.Fatalf("get-body can't parse: %v\n", whichArg)
	}
	mbm := nbodygrpc.ModBodyMessage{
		Id:    int64(id),
		Name:  name,
		Class: "",
	}
	b, err := c.nbodyCli.GetBody(context.Background(), &mbm)
	if err != nil {
		log.Fatalf("get-body error: %v\n", err)
	}
	result := "" +
		"id: %v\n" +
		"x,y,z: %v,%v,%v\n" +
		"vx,vy,vz: %v,%v,%v\n" +
		"mass: %v\n" +
		"radius: %v\n" +
		"is-sun: %v\n" +
		"intensity: %v\n" +
		"collision: %v\n" +
		"color: %v\n" +
		"frag-factor: %v\n" +
		"frag-step: %v\n" +
		"telemetry: %v\n" +
		"name: %v\n" +
		"class: %v\n"
	fmt.Printf(result,
		b.Id,
		b.X, b.Y, b.Z,
		b.Vx, b.Vy, b.Vz,
		b.Mass,
		b.Radius,
		b.IsSun,
		b.Intensity,
		b.CollisionBehavior,
		b.BodyColor,
		b.FragFactor, b.FragStep,
		b.WithTelemetry,
		b.Name,
		b.Class)
}

//
// Adds one or more bodies to the sim on the server depending on the passed command
//
// args:
//   cmd  Either 'add-body' or 'add-bodies'
//   args Everything on the command line after 'cmd' E.g.:
//        400 400 -400 -850000000 923000000 -350000000 9E5 5 color=red qty=600 delay=.3
//
func (c *client) addBodies(cmd string, args []string) {
	const firstNonPositional = 8
	p := [8]float64{}
	for i := 0; i < 8; i++ {
		p[i] = parseFloatOrPanic(args[i])
	}
	isSun, withTelemetry, pinned := false, false, false
	collisionBehavior := globals.Elastic
	color := globals.Random
	fragFactor, fragStep, intensity := float64(0), float64(0), float64(100)
	name, class := "", ""

	// only for add-bodies, not add-body:
	qty := 1
	delay, positionRandom, velocityRandom, massRandom, radiusRandom := float64(0), float64(0), float64(0),
		float64(0), float64(0)

	for i := firstNonPositional; i < len(args); i++ {
		nv := strings.Split(args[i], "=")
		switch strings.ToLower(nv[0]) {
		case "is-sun":
			isSun = true
		case "intensity":
			intensity = parseFloatOrPanic(nv[1])
		case "collision":
			collisionBehavior = globals.ParseCollisionBehavior(nv[1])
		case "color":
			color = globals.ParseBodyColor(nv[1])
		case "frag-factor":
			fragFactor = parseFloatOrPanic(nv[1])
		case "frag-step":
			fragStep = parseFloatOrPanic(nv[1])
		case "telemetry":
			withTelemetry = true
		case "name":
			name = nv[1]
		case "class":
			class = nv[1]
		case "pinned":
			pinned = true
		case "qty":
			qty = parseIntOrPanic(nv[1])
		case "delay":
			delay = parseFloatOrPanic(nv[1])
		case "posrand":
			positionRandom = parseFloatOrPanic(nv[1])
		case "vrand":
			velocityRandom = parseFloatOrPanic(nv[1])
		case "massrand":
			massRandom = parseFloatOrPanic(nv[1])
		case "rrand":
			radiusRandom = parseFloatOrPanic(nv[1])
		default:
			panic("Unknown param: " + args[i])
		}
	}
	if cmd == "add-body" {
		c.addOneBody(p[0], p[1], p[2], p[3], p[4], p[5], p[6], p[7], isSun, intensity, collisionBehavior, color,
			fragFactor, fragStep, withTelemetry, name, class, pinned)
	} else { //add-bodies
		c.addMultiBodies(p[0], p[1], p[2], p[3], p[4], p[5], p[6], p[7], isSun, intensity, collisionBehavior, color,
			fragFactor, fragStep, withTelemetry, name, class, pinned, qty,
			delay, positionRandom, velocityRandom, massRandom, radiusRandom)
	}
}

//
// Adds one body to the simulation on the server with the passed params
//
func (c *client) addOneBody(x, y, z, vx, vy, vz, mass, radius float64,
	isSun bool, intensity float64,
	collisionBehavior globals.CollisionBehavior, color globals.BodyColor,
	fragFactor, fragStep float64,
	withTelemetry bool,
	name, class string,
	pinned bool) {

	if isSun && color == globals.Random {
		color = globals.White
	}

	bd := nbodygrpc.BodyDescription{
		Id:                0, // ignored on add
		X:                 x,
		Y:                 y,
		Z:                 z,
		Vx:                vx,
		Vy:                vy,
		Vz:                vz,
		Mass:              mass,
		Radius:            radius,
		IsSun:             isSun,
		Intensity:         intensity,
		CollisionBehavior: grpcserver.SimCbToGrpcCb(collisionBehavior),
		BodyColor:         grpcserver.SimColorToGrpcColor(color),
		FragFactor:        fragFactor,
		FragStep:          fragStep,
		WithTelemetry:     withTelemetry,
		Name:              name,
		Class:             class,
		Pinned:            pinned,
	}
	result, _ := c.nbodyCli.AddBody(context.Background(), &bd)
	println(result.Message)
}

//
// Adds multiple bodies to the simulation on the server with the passed params. The delay, and ...random
// args randomize how the bodies are added
//
func (c *client) addMultiBodies(x, y, z, vx, vy, vz, mass, radius float64,
	isSun bool, intensity float64,
	collisionBehavior globals.CollisionBehavior, color globals.BodyColor,
	fragFactor, fragStep float64,
	withTelemetry bool,
	name, class string,
	pinned bool,
	qty int,
	delay, positionRandom, velocityRandom, massRandom, radiusRandom float64) {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < qty; i++ {
		wx := randomize(x, positionRandom)
		wy := randomize(y, positionRandom)
		wz := randomize(z, positionRandom)
		wvx := randomize(vx, velocityRandom)
		wvy := randomize(vy, velocityRandom)
		wvz := randomize(vz, velocityRandom)
		wmass := randomize(mass, massRandom)
		wradius := randomize(radius, radiusRandom)
		c.addOneBody(wx, wy, wz, wvx, wvy, wvz, wmass, wradius, isSun, intensity, collisionBehavior, color, fragFactor, fragStep,
			withTelemetry, name, class, pinned)
		time.Sleep(time.Millisecond * time.Duration(1000*delay))
	}
}

//
// Modifies one or more bodies in the sim on the server using the passed params
//
// args:
//   whichArg  Specifies which body or bodies to modify. Either 'id=', 'class=', or 'name='. E.g.: mod-body id=123.
//             Or: mod-body name=jupiter or mod-body class=asteroid
//
//
func (c *client) modBodies(whichArg string, args []string) {
	which := strings.Split(whichArg, "=")
	id := -1
	name := ""
	class := ""
	var err error
	if len(which) != 2 {
		log.Fatalf("mod-body needs id=n, class=foo, or name=foo\n")
	}
	if which[0] == "id" {
		id, err = strconv.Atoi(which[1])
	} else if which[0] == "name" {
		name = which[1]
	} else if which[0] == "class" {
		class = which[1]
	} else {
		log.Fatalf("mod-body needs id=n, class=foo, or name=foo\n")
	}
	if err != nil {
		log.Fatalf("mod-body can't parse: %v\n", whichArg)
	}
	validMods := map[string]interface{}{"x": nil, "y": nil, "z": nil, "vx": nil, "vy": nil, "vz": nil, "mass": nil,
		"radius": nil, "sun": nil, "intensity": nil, "collision": nil, "color": nil, "frag-factor": nil, "frag-step": nil,
		"telemetry": nil, "exists": nil}
	for _, arg := range args {
		p := strings.Split(arg, "=")
		_, ok := validMods[p[0]]
		if !ok {
			log.Fatalf("mod-body unknown property: %v\n", arg)
		}
		if p[0] == "sun" || p[0] == "intensity" {
			log.Fatalf("option not implemented yet: %v\n", arg)
		}
	}
	mb := nbodygrpc.ModBodyMessage{
		Id:    int64(id),
		Name:  name,
		Class: class,
		P:     args,
	}
	result, _ := c.nbodyCli.ModBody(context.Background(), &mb)
	println(result.Message)
}

//
// gRPC client entry point
//
func main() {
	if len(os.Args) == 1 {
		basicHelp()
		return
	}
	cmd := strings.ToLower(os.Args[1])
	validateArgs(cmd, len(os.Args[1:]))

	client := newClient()
	defer func() {
		client.stopClient()
		if r := recover(); r != nil {
			log.Fatalf("Error: %v\n", r)
		}
	}()

	switch cmd {
	case "set-threads":
		client.setComputationThreads(os.Args[2])
	case "set-queue-size":
		client.setResultQueueSize(os.Args[2])
	case "set-time-scale":
		client.setSmoothing(os.Args[2])
	case "set-restitution":
		client.setRestitutionCoefficient(os.Args[2])
	case "remove-bodies":
		client.removeBodies(os.Args[2])
	case "mod-body":
		fallthrough
	case "mod-bodies":
		client.modBodies(os.Args[2], os.Args[3:])
	case "get-config":
		client.getCurrentConfig()
	case "get-body":
		client.getBody(os.Args[2])
	case "add-body":
		fallthrough
	case "add-bodies":
		client.addBodies(os.Args[1], os.Args[2:])
	default:
		println("Unsupported cmd: " + os.Args[1])
	}
}

//
// Do some basic cmdline arg validation
//
// args:
//   cmd    the command, e.g. 'set-threads'
//   argCnt the count of args from cmd forward, e.g. if 'cmd' is 'set-threads' and the cmd line
//          is ["set-threads", "1"] then argCnt = 2
//
func validateArgs(cmd string, argCount int) {
	switch cmd {
	case "set-threads":
		fallthrough
	case "set-queue-size":
		fallthrough
	case "set-time-scale":
		fallthrough
	case "set-restitution":
		fallthrough
	case "get-body":
		fallthrough
	case "remove-bodies":
		if argCount != 2 {
			log.Fatalf("Missing value for cmd: " + cmd)
		}
	case "mod-body":
		fallthrough
	case "mod-bodies":
		if argCount < 3 {
			log.Fatalf("Need at least two params for cmd: " + cmd)
		}
	case "add-body":
		fallthrough
	case "add-bodies":
		if argCount < 9 {
			log.Fatalf("Eight positional params (x, y, z, vx, vy, vz, mass, radius) minimally required for cmd: " + cmd)
		}
	}
}

//
// Prints some basic help
//
func basicHelp() {
	help := "N-Body Golang:\n\nSupported commands:\n\n" + "" +
		" set-threads <threads>\n" +
		" set-queue-size <size>\n" +
		" set-time-scale <timescale>\n" +
		" set-restitution <coefficient>\n" +
		" remove-bodies <count>\n" +
		" mod-body <id= or name= or class=> <property=value> ...\n" +
		" mod-bodies (same as above)\n" +
		" get-config (no args)\n" +
		" get-body <id= or name=>\n" +
		" add-body x, y, z, vx, vy, vz, mass, radius, <other-property=value> ...\n" +
		" add-bodies (same as above plus qty= delay= etc.)\n\n" +
		"see: https://github.com/aceeric/nbodyjava for full documentation (this is a port from a Java project)\n"
	println(help)
}

//
// Parses and returns the passed string as an int, or panics
//
func parseIntOrPanic(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic("Not an int: " + s)
	}
	return i
}

//
// Parses and returns the passed string as a float, or panics
//
func parseFloatOrPanic(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic("Not a float: " + s)
	}
	return f
}

//
// if arg 'random' is non-zero then it is randomized added to 'val' before returning
// 'val'. If 'random' is zero, then 'val' is returned un-modified. Used to randomize
// body properties when adding. E.g.
//    add-bodies x=100 ... posrand=10
//
// The result of this would be to randomize x (and y and z) between 100 and 110
//
func randomize(val float64, random float64) float64 {
	if random == 0 {
		return val
	}
	return val + (rand.Float64() * random)
}
