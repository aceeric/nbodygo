package main

import (
	"container/list"
	"nbodygo/cmd/body"
	"nbodygo/cmd/globals"
	"nbodygo/cmd/logging"
	"nbodygo/cmd/sim"
	"nbodygo/internal/g3n/math32"
	"os"
	"regexp"
	"strconv"
	"strings"
)

/*
Entry point for the server
*/

const (
	// if no sim name is provided on the command line, run this sim
	defaultSimName = "Sim1"

	// empty sim starts everything up with on bodies: --sim-name=Empty
	emptySimName = "Empty"
)

//
// state for the main function - initialized with defaults that can be overridden from the
// command line
//
var vars = struct {
	resolution               [2]int
	render                   bool
	workers                  int
	scaling                  float64
	simName                  string
	defaultCollisionBehavior globals.CollisionBehavior
	bodyCount                int
	csvPath                  string
	defaultBodyColor         globals.BodyColor
	initialCam               *math32.Vector3
	simArgs                  string
	vSync                    bool
	frameRate                int
	runMillis                int
}{
	resolution:               [2]int{2560, 1405},
	render:                   true,
	workers:                  5,
	scaling:                  .000000001,
	simName:                  "",
	defaultCollisionBehavior: globals.Elastic,
	bodyCount:                1000,
	csvPath:                  "",
	defaultBodyColor:         globals.Random,
	initialCam:               math32.NewVector3(-100, 300, 1200),
	simArgs:                  "",
	vSync:                    false, // not currently supported
	frameRate:                0,     // "
	runMillis:                -1,    // run forever in --no-render mode
}

//
// Parses the command line, creates a sim based on command line args, and launches the sim
//
func main() {
	if !parseArgs() {
		return
	}
	// initialize a list of bodies representing the simulation, and optionally a worker function that
	// will modify the body collection concurrently while the sim is running
	var bodies []*body.Body
	var simWorker sim.Worker

	if len(vars.csvPath) > 0 {
		bodies = sim.FromCsv(vars.csvPath, vars.bodyCount, vars.defaultCollisionBehavior, vars.defaultBodyColor)
	} else if strings.EqualFold(vars.simName, emptySimName) {
		bodies = []*body.Body{}
	} else {
		bodies, simWorker = sim.Generate(vars.simName, vars.bodyCount, vars.defaultCollisionBehavior,
			vars.defaultBodyColor, vars.simArgs)
		if bodies == nil {
			println("ERROR: could not build sim specified on the command line: " + vars.simName)
			return
		}
	}
	logging.InitializeLogging()
	// initialize and run the simulation
	sim.NewNBodySimBuilder().
		Bodies(bodies).
		Workers(vars.workers).
		Scaling(vars.scaling).
		InitialCam(*vars.initialCam).
		SimWorker(simWorker).
		Render(vars.render).
		Resolution(vars.resolution).
		VSync(vars.vSync).
		FrameRate(vars.frameRate).
		RunMillis(vars.runMillis).
		Build().
		Run()
}

//
// A very rudimentary command-line option parser. Accepts short-form opts like -t and long-form like --threads.
// Accepts this form: -t 1 and --threads 1, as well as this form -t=1 and --threads=1. Does not accept
// concatenated short form opts in cases where such opts don't accept params. E.g. doesn't handle: -ot=1 where
// -o is a parameterless option, and -t takes a value (one in this example.) Doesn't have great error handling
// so - is fragile with respect to parsing errors.
//
// Sets values in the 'vars' struct corresponding to command line args.
//
// returns false if there was an arg parse error, else return true
//
func parseArgs() bool {
	argQueue := list.New()
	for i := 1; i < len(os.Args); i++ {
		// handle --opt=value
		s := strings.Split(os.Args[i], "=")
		argQueue.PushBack(s[0])
		if len(s) == 2 {
			argQueue.PushBack(s[1])
		}
	}
	// define a 'next arg' function to treat the command line like a fifo queue of strings
	nextArg := func() *string {
		var argval string
		if a := argQueue.Front(); a != nil {
			argval = a.Value.(string)
			argQueue.Remove(argQueue.Front())
			return &argval
		}
		return nil
	}
	for arg := nextArg(); arg != nil; arg = nextArg() {
		switch strings.ToLower(*arg) {
		case "-z":
			fallthrough
		case "--resolution":
			s := nextArg()
			sSplit := regexp.MustCompile("[xX]").Split(*s, 2)
			if len(sSplit) != 2 {
				println("Invalid resolution: " + *s)
				return false
			}
			z, _ := strconv.ParseInt(sSplit[0], 0, 32)
			vars.resolution[0] = int(z)
			z, _ = strconv.ParseInt(sSplit[1], 0, 32)
			vars.resolution[1] = int(z)
			break
		case "--vsync":
			vars.vSync, _ = strconv.ParseBool(*nextArg())
			println("--vsync is currently unimplemented")
			return false
		case "--frame-rate":
			z, _ := strconv.ParseInt(*nextArg(), 0, 32)
			vars.frameRate = int(z)
			println("--frame-rate is currently unimplemented")
			return false
		case "--run-millis":
			fallthrough
		case "-u":
			z, _ := strconv.ParseInt(*nextArg(), 0, 32)
			vars.runMillis = int(z)
			break
		case "-r":
			fallthrough
		case "--no-render":
			vars.render = false
			break
		case "-n":
			fallthrough
		case "--sim-name":
			vars.simName = strings.Title(*nextArg())
			break
		case "-a":
			fallthrough
		case "--sim-args":
			vars.simArgs = *nextArg()
			break
		case "-c":
			fallthrough
		case "--collision":
			vars.defaultCollisionBehavior = globals.ParseCollisionBehavior(*nextArg())
			break
		case "-b":
			fallthrough
		case "--bodies":
			z, _ := strconv.ParseInt(*nextArg(), 0, 32)
			vars.bodyCount = int(z)
			break
		case "-t":
			fallthrough
		case "--threads":
			z, _ := strconv.ParseInt(*nextArg(), 0, 32)
			vars.workers = int(z)
			break
		case "-m":
			fallthrough
		case "--scaling":
			vars.scaling, _ = strconv.ParseFloat(*nextArg(), 32)
			break
		case "-f":
			fallthrough
		case "--csv":
			vars.csvPath = *nextArg()
			break
		case "-l":
			fallthrough
		case "--body-color":
			vars.defaultBodyColor = globals.ParseBodyColor(*nextArg())
			break
		case "-i":
			fallthrough
		case "--initial-cam":
			vars.initialCam = globals.ParseVector(*nextArg())
			break
		case "-h":
			fallthrough
		case "--help":
			println("Sorry: help not implemented yet...")
			return false
		default:
			println("ERROR: unknown option: " + *arg)
			return false
		}
	}
	if len(vars.simName) != 0 && len(vars.csvPath) != 0 {
		println("ERROR: provide *either* a sim name *or* a csv path, but not both")
		return false
	}
	if len(vars.simName) == 0 && len(vars.csvPath) == 0 {
		vars.simName = defaultSimName
	}
	return true
}
