package sim

import (
	"math/rand"
	"nbodygo/cmd/body"
	"nbodygo/cmd/globals"
	"nbodygo/cmd/interfaces"
	"nbodygo/cmd/util"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type GeneratorFunc = func(int, globals.CollisionBehavior, globals.BodyColor, string) (
	[]interfaces.SimBody, func(interfaces.SimBodyCollection))

type WorkerFunc = func(interfaces.SimBodyCollection)

type Generator interface {}
type generator struct {}
const (
	solarMass = 1.98892e30
)

var Instance Generator = generator{}

func Generate(simName string, bodyCount int, collisionBehavior globals.CollisionBehavior,
	defaultBodyColor globals.BodyColor, simArgs string) ([]interfaces.SimBody, WorkerFunc) {
	value := reflect.ValueOf(Instance)
	ptr := reflect.New(reflect.TypeOf(Instance))
	temp := ptr.Elem()
	temp.Set(value)
	method := value.MethodByName(simName)
	if !method.IsValid() {
		return nil, nil
	}
	params := []reflect.Value{
		reflect.ValueOf(bodyCount),
		reflect.ValueOf(collisionBehavior),
		reflect.ValueOf(defaultBodyColor),
		reflect.ValueOf(simArgs),
	}
	retVals := method.Call(params)
	bodies := retVals[0].Interface().([]interfaces.SimBody)
	workerFunc := retVals[1].Interface().(WorkerFunc)
	return bodies, workerFunc
}

func (g generator) Sim1(bodyCount int, collisionBehavior globals.CollisionBehavior, defaultBodyColor globals.BodyColor,
	simArgs string) ([]interfaces.SimBody, WorkerFunc) {
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
	var bodies []interfaces.SimBody
	id := 0
	var vx, vy, vz, y, mass, radius float64
	V := float64(958000000)
	rand.Seed(time.Now().UnixNano())
	for i := -1; i <= 1; i += 2 {
		for j := -1; j <= 1; j += 2 {
			xc := dist * float64(i)
			zc := dist * float64(j)
			color := defaultBodyColor

			if        i == -1 && j == -1 {vx = -V; vz =  V; y = +100; if defaultBodyColor == globals.Random { color = globals.Red}
			} else if i == -1 && j ==  1 {vx =  V; vz =  V; y = -100; if defaultBodyColor == globals.Random { color = globals.Yellow}
			} else if i ==  1 && j ==  1 {vx =  V; vz = -V; y = +100; if defaultBodyColor == globals.Random { color = globals.Lightgray}
			} else                       {vx = -V; vz = -V; y = -100; if defaultBodyColor == globals.Random { color = globals.Cyan}}

			for c := 0; c < bodyCount / 4; c++ {
				vy = .5 - rand.Float64()
				f := rand.Float64()
				if float64(c) < float64(bodyCount) * .0025 {
					radius = 8 * f
				} else {
					radius = 3 * f
				}
				mass = radius * solarMass * .000005
				v := util.GetVectorEven(*util.NewVector3(xc, y, zc), clumpRadius)
				b := body.NewBody(id, v.X, v.Y, v.Z, vx, vy, vz, mass, radius, collisionBehavior,
					color, 0, 0, false, "", "",false)
				bodies = append(bodies, &b)
				id++
			}
		}
	}
	return createSunAndAddToList(bodies, id, 0, 0, 0, 25 * solarMass * .11, 35), nil
}

//
// Creates a sun body with larger mass, very low (non-zero) velocity, placed at 0, 0, 0 and
// places it into the passed body list. Sun bodies are light sources. Each sim needs at least one light
// source.
//
// args:
//   bodies            - array of bodies to add the sun into
//   id                - if of the sun
//   x,y,z,mass,radius - core body properties
//
// returns:
//   the passed array with the sun appended
//
func createSunAndAddToList(bodies []interfaces.SimBody, id int, x, y, z, mass, radius float64) []interfaces.SimBody {
	b := body.NewBody(id, x, y, z, -3, -3, -5, mass, radius, globals.Subsume,
		globals.White, 0, 0, false, "the-sun", "",false)
	b.SetSun()
	bodies = append(bodies, &b)
	return bodies
}