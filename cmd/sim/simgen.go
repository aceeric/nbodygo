package sim

/*
Contains functions to generate canned simulations
*/

import (
	"math/rand"
	"nbodygo/cmd/body"
	"nbodygo/cmd/globals"
	"nbodygo/cmd/util"
	"reflect"
	"strconv"
	"strings"
	"time"
)

//
// Some sims start a worker to inject bodies into the simulation after the sim is started, and return the
// function so it can be called by the simulation runner (the Run function in the 'sim' package.) This defines
// the function signature for those workers.
//
type SimWorker = func(*body.BodyCollection)

//
// These declarations support reflection, which is used to get the function corresponding to the --sim-name
// command-line option
//
type Generator interface{}
type generator struct{}
var instance Generator = generator{}

const (
	solarMass = 1.98892e30
)

//
// Generates and returns a simulation. The args are passed directly through to the sim generation function
// and might be ignored by that function if not applicable.
//
// args:
//  simName           The name of the sim. This must be a function name in this package. E.g. "Sim1". The
//                    matching function name is obtained using reflection
//  bodyCount         The number of bodies
//  collisionBehavior collision behavior if not explicitly defined by the sim
//	defaultBodyColor  " body color
//	simArgs           Some sims take args to customize their behavior. Refer to the individual sim functions
//                    for specifics
//
// returns: a list of bodies, and optionally a worker function to modify the sim after it starts
//
func Generate(simName string, bodyCount int, collisionBehavior globals.CollisionBehavior,
	defaultBodyColor globals.BodyColor, simArgs string) ([]*body.Body, SimWorker) {

	// use reflection to get the sim name passed in the 'simName' arg
	value := reflect.ValueOf(instance)
	ptr := reflect.New(reflect.TypeOf(instance))
	temp := ptr.Elem()
	temp.Set(value)
	method := value.MethodByName(simName)
	if !method.IsValid() {
		return nil, nil
	}

	// initialize sim function parameters
	params := []reflect.Value{
		reflect.ValueOf(bodyCount),
		reflect.ValueOf(collisionBehavior),
		reflect.ValueOf(defaultBodyColor),
		reflect.ValueOf(simArgs),
	}
	// call the sim generator function
	retVals := method.Call(params)
	// get the return values
	bodies := retVals[0].Interface().([]*body.Body)
	workerFunc := retVals[1].Interface().(SimWorker)
	return bodies, workerFunc
}

//
// Creates four clumps of bodies centered at "left, right, front, and back", with each clump organized
// spherically, around that center point. The velocity of the bodies in each clump is set so
// that each clump will be captured by the sun. Each clump contains mostly small, similar-sized
// bodies but also a few larger bodies are included for variety.
//
// args:
//  simArgs CSV in the form: radius of clump, distance of clump from center. E.g.: "30,200"
//          (these are the defaults if no arg provided)
//
// returns: a list of bodies, no worker function

func (g generator) Sim1(bodyCount int, collisionBehavior globals.CollisionBehavior, defaultBodyColor globals.BodyColor,
	simArgs string) ([]*body.Body, SimWorker) {
	var parsedSimArgs []string
	clumpRadius := float64(30)
	dist := float64(200)
	if len(simArgs) == 0 {
		parsedSimArgs = []string{"30", "200"}
	} else {
		parsedSimArgs = strings.Split(simArgs, ",")
	}
	if len(parsedSimArgs) > 0 {
		z, _ := strconv.ParseFloat(parsedSimArgs[0], 32)
		clumpRadius = z
	}
	if len(parsedSimArgs) > 1 {
		z, _ := strconv.ParseFloat(parsedSimArgs[1], 32)
		dist = z
	}
	var bodies []*body.Body
	var vx, vy, vz, y, mass, radius float64
	V := float64(958000000)
	rand.Seed(time.Now().UnixNano())
	for i := -1; i <= 1; i += 2 {
		for j := -1; j <= 1; j += 2 {
			xc := dist * float64(i)
			zc := dist * float64(j)
			color := defaultBodyColor

			if i == -1 && j == -1 {
				vx = -V
				vz = V
				y = +100
				if defaultBodyColor == globals.Random {
					color = globals.Red
				}
			} else if i == -1 && j == 1 {
				vx = V
				vz = V
				y = -100
				if defaultBodyColor == globals.Random {
					color = globals.Yellow
				}
			} else if i == 1 && j == 1 {
				vx = V
				vz = -V
				y = +100
				if defaultBodyColor == globals.Random {
					color = globals.Lightgray
				}
			} else {
				vx = -V
				vz = -V
				y = -100
				if defaultBodyColor == globals.Random {
					color = globals.Cyan
				}
			}

			for c := 0; c < bodyCount/4; c++ {
				vy = .5 - rand.Float64()
				f := rand.Float64()
				if float64(c) < float64(bodyCount)*.0025 {
					radius = 8 * f
				} else {
					radius = 3 * f
				}
				mass = radius * solarMass * .000005
				v := util.GetVectorEven(*util.NewVector3(xc, y, zc), clumpRadius)
				b := body.NewBody(body.NextId(), v.X, v.Y, v.Z, vx, vy, vz, mass, radius, collisionBehavior,
					color, 0, 0, false, "", "", false)
				bodies = append(bodies, b)
			}
		}
	}
	return createSunAndAddToList(bodies, body.NextId(), 0, 0, 0, 25*solarMass*.11, 35, 100), nil
}

//
// Generates a sun at 0,0,0 and a cluster of bodies off-screen headed for a very close pass around
// with the sun at high velocity. Typically, a few bodies are captured by the sun but most travel away
//
// args:
//  simArgs Unused by this generator
//
// returns: a list of bodies, no worker function
//
func (g generator) Sim2(bodyCount int, collisionBehavior globals.CollisionBehavior, defaultBodyColor globals.BodyColor,
	simArgs string) ([]*body.Body, SimWorker) {
	_ = simArgs
	bodies := createSunAndAddToList([]*body.Body{}, body.NextId(), 0, 0, 0, 25*solarMass*.1, 25, 100)
	rand.Seed(time.Now().UnixNano())
	for i := 1; i < bodyCount; i++ {
		v := util.GetVectorEven(*util.NewVector3(500, 500, 500), 50)
		mass := rand.Float64() * solarMass * .000005
		radius := rand.Float64() * 4
		b := body.NewBody(body.NextId(), v.X, v.Y, v.Z, -1124500000, -824500000, -1124500000,
			mass, radius, collisionBehavior, defaultBodyColor, 1, 1, false, "",
			"", false)
		bodies = append(bodies, b)
	}
	return bodies, nil
}

//
// Generates a simulation with a sun far removed from the focus area just to serve as light source. Creates
// two clusters composed of many colliding spheres in close proximity. The two clusters exert gravitational
// attraction toward each other as if they were solids. They also exert gravitational force within themselves,
// preserving their spherical shape. The two clusters orbit and then collide, merging into a single cluster
// of colliding spheres.
//
// After the sim starts, the worker function returned by the generator injects a series of bodies
// gradually into the simulation. The additional bodies come in over a period of a few minutes.
//
// This sim is dependent on body count - I run it with ~1000 bodies. Fewer, and the attraction isn't enough
// to bring the clusters together. More, and the two clusters merge too soon. This sim should be run with
// elastic collision. This example was useful to surface some subtleties with regard to how the simulation
// handles lots of concurrent elastic collisions
//
// args:
//   simArgs CSV in the form: radius of clump, mass of bodies in clump, body count to inject. E.g.:,
//           "70,90E25,500" (these are the defaults if no arg provided)
//
// returns: a list of bodies, and a worker function to insert additional bodies
//
func (g generator) Sim3(bodyCount int, collisionBehavior globals.CollisionBehavior, defaultBodyColor globals.BodyColor,
	simArgs string) ([]*body.Body, SimWorker) {
	var parsedSimArgs []string
	if len(simArgs) == 0 {
		parsedSimArgs = []string{"50", "90E25", "700"}
	} else {
		parsedSimArgs = strings.Split(simArgs, ",")
	}
	radius, mass := float64(50), 90E25
	injectCnt := 700

	if len(parsedSimArgs) > 0 {
		radius, _ = strconv.ParseFloat(parsedSimArgs[0], 64)
	}
	if len(parsedSimArgs) > 1 {
		mass, _ = strconv.ParseFloat(parsedSimArgs[1], 64)
	}
	if len(parsedSimArgs) > 2 {
		z, _ := strconv.ParseInt(parsedSimArgs[2], 0, 32)
		injectCnt = int(z)
	}
	bodies := createSunAndAddToList([]*body.Body{}, body.NextId(), 100000, 100000, 100000, 1, 500, 4E5)
	for j := -1; j <= 1; j += 2 {
		for i := 0; i < bodyCount/2; i++ {
			bodyColor := defaultBodyColor
			if defaultBodyColor == globals.Random {
				if j == 01 {
					bodyColor = globals.Yellow
				} else {
					bodyColor = globals.Red
				}
			}
			v := util.GetVectorEven(*util.NewVector3(float64(j)*70, float64(j)*70, float64(j)*70), radius)
			b := body.NewBody(body.NextId(), v.X, v.Y, v.Z, float64(j)*121185000, float64(j)*121185000, float64(j)*-121185000,
				mass, 5, collisionBehavior, bodyColor, 1, 1, false, "",
				"", false)
			bodies = append(bodies, b)
		}
	}
	simWorker := func(bc *body.BodyCollection) {
		cnt := 0
		rand.Seed(time.Now().UnixNano())
		for {
			if cnt >= injectCnt {
				return
			}
			cnt++

			x := rand.Float64()*5 - 200
			y := rand.Float64()*5 + 400
			z := rand.Float64()*5 - 200
			radius := rand.Float64() * 5
			mass := radius * 2.93E+12
			bodyColor := defaultBodyColor
			if defaultBodyColor == globals.Random {
				bodyColor = globals.Blue
			}
			b := body.NewBody(body.NextId(), x, y, z, -99827312, 112344240, 323464000,
				mass, radius, collisionBehavior, bodyColor, 1, 1, false, "",
				"", false)
			bc.Enqueue(body.NewAdd(b))
			time.Sleep(time.Millisecond * 500)
		}
	}
	return bodies, simWorker
}

// Generates a sun and a line of bodies along the x axis all in the same plane moving at the same
// velocity showing that nearer objects are captured more quickly than farther objects
//
//  simArgs Unused by this generator
//
// returns:  a list of bodies, and no sim worker

func (g generator) Sim4(bodyCount int, collisionBehavior globals.CollisionBehavior, defaultBodyColor globals.BodyColor,
	simArgs string) ([]*body.Body, SimWorker) {
	_= simArgs
	bodies := createSunAndAddToList([]*body.Body{}, body.NextId(), 0, 0, 0, solarMass, 30, 90)
	for i := 1; i < bodyCount; i++ {
		mass := 9e5
		radius := float64(2)
		b := body.NewBody(body.NextId(), float64(i * 4) + 100, 0, 0, // x,y,z
			0, 0, -824500000 + float64(i * 1E6), // vx,vy,vz
			mass, radius, collisionBehavior, defaultBodyColor, 1, 1, false,
			"", "", false)
		bodies = append(bodies, b)
	}
	return bodies, nil
}

//
// Generates a sun far removed from the focus area just to serve as light source. Creates a large
// planet at the center of the sim orbited by three moons. Creates a small impactor headed for
// the large planet. The impactor is configured to fragment into many smaller bodies on impact.
//
// After the sim starts, the worker returned by the method monitors the simulation and when the impact
// occurs it changes the planet's collision behavior from ELASTIC to SUBSUME. As a result, any of the
// smaller fragments that subsequently strike the planet are absorbed into the planet.
//
// args:
//   bodyCount, bodyCount, defaultBodyColor are ignored
//   simArgs CSV configuring the impactor in the form frag factor,frag step. E.g.:
//           .01F,1000 (the default). The larger the factor, the more force is required to
//           cause fragmentation. The larger the step, the more fragments are created.
//
// returns a body list and a worker as described
//
func (g generator) Sim5(bodyCount int, collisionBehavior globals.CollisionBehavior, defaultBodyColor globals.BodyColor,
	simArgs string) ([]*body.Body, SimWorker) {
	_ = bodyCount
	_ = collisionBehavior
	_ = defaultBodyColor

	fragFactor, fragStep := .01, float64(1000)
	if len(simArgs) > 0 {
		parsedSimArgs := strings.Split(simArgs, ",")
		if len(parsedSimArgs) >= 1 {
			fragFactor, _ = strconv.ParseFloat(parsedSimArgs[0], 64)
		}
		if len(parsedSimArgs) >= 2 {
			fragStep, _ = strconv.ParseFloat(parsedSimArgs[1], 64)
		}
	}
	bodies := createSunAndAddToList([]*body.Body{}, body.NextId(), 100000, 100000, 001000, 1, 500, 4E5)

	// planet
	planet := body.NewBody(body.NextId(), 0, 0, 0, 12, 12, 12, 9E30, 145, globals.Elastic,
		globals.Red, 0, 0, false, "", "", false)
	bodies = append(bodies, planet)

	// moons
	m1 := body.NewBody(body.NextId(), 50, 0, -420, -980000000, 12, -500000000, 9E20, 35, globals.Subsume,
		globals.Lightgray, 0, 0, false, "", "", false)
	bodies = append(bodies, m1)

	m2 := body.NewBody(body.NextId(), -400, 50, 405, 530000000, -313000000, 520000000, 9E19, 5, globals.Elastic,
		globals.Blue, 0, 0, false, "", "", false)
	bodies = append(bodies, m2)

	m3 := body.NewBody(body.NextId(), 70, 0, -520, -880000000, -10000, -300000000 , 11E22, 15, globals.Elastic,
		globals.Green, 0, 0, false, "", "", false)
	bodies = append(bodies, m3)

	// impactor
	im := body.NewBody(body.NextId(), 900, -900, 900, -450000000, 723000000, -350000000, 9E12, 10, globals.Fragment,
		globals.Yellow, fragFactor, fragStep, false, "", "", false)
	bodies = append(bodies, im)

	simWorker := func(bc *body.BodyCollection) {
		for {
			if bc.Count() > 6 {
				bc.ModBody(planet.Id, "", "",  []string{"collision=subsume"})
				return
			}
			time.Sleep(time.Millisecond * 1000)
		}
	}
	return bodies, simWorker
}

//
// Validate collision resolution using the deferred approach
//
func (g generator) SimTest(bodyCount int, collisionBehavior globals.CollisionBehavior, defaultBodyColor globals.BodyColor,
	simArgs string) ([]*body.Body, SimWorker) {
	_ = bodyCount
	_ = defaultBodyColor
	_ = simArgs
	bodies := createSunAndAddToList([]*body.Body{}, body.NextId(), 20000, 20000, 20000, 1, 500, 10000)

	// 0,0,0 stationary
	m1 := body.NewBody(body.NextId(), 0, 0, 0, 0, 0, 0, 9E29, 60, collisionBehavior,
		globals.Red, 0, 0, false, "", "", false)
	bodies = append(bodies, m1)

	// left heading down right
	m2 := body.NewBody(body.NextId(), -350, 350, 0, 530000000, -500000000, 0, 9E29, 60,
		collisionBehavior, globals.Green, 0, 0, false, "", "",
		false)
	bodies = append(bodies, m2)

	// right heading down left
	m3 := body.NewBody(body.NextId(), 350, 350, 0, -530000000, -500000000, 0, 9E29, 60,
		collisionBehavior, globals.Yellow, 0, 0, false, "", "",
		false)
	bodies = append(bodies, m3)

	return bodies, nil
}

//
// Creates a sun body with very low (non-zero) velocity, placed at the passed coordinates and
// places it into the passed body list. Sun bodies are light sources. Each sim needs at
// least one light source.
//
// args:
//   bodies            - array of bodies to add the sun into
//   id                - id of the sun
//   x,y,z,mass,radius - core body properties
//   intensity         - light source intensity. For a sun far removed from the focus area, increase this to
//                       offset the decay
//
// returns:
//   the passed array with the sun appended
//
func createSunAndAddToList(bodies []*body.Body, id int, x, y, z, mass, radius float64, intensity float64) []*body.Body {
	b := body.NewBody(id, x, y, z, -3, -3, -5, mass, radius, globals.Subsume,
		globals.White, 0, 0, false, "the-sun", "", true)
	b.SetSun(intensity)
	return append(bodies, b)
}
