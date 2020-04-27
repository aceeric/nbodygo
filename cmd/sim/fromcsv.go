package sim

import (
	"bufio"
	"encoding/csv"
	"io"
	"log"
	"nbodygo/cmd/body"
	"nbodygo/cmd/globals"
	"os"
	"strconv"
	"strings"
)

//
// Parses a CSV file into a list of bodies. The format must be comma-delimited, with fields:
//
//   x, y, z, vx, vy, vz, mass, radius, is_sun, collision_behavior, color, fragmentation_factor, fragmentation_step
//
// Everything from 'x' through 'radius' is required - and is parsed as a float. Everything else is optional.
// Comments are allowed: any line where '#' is the first character.  If 'is_sun' is omitted, the loader defaults
// it to 'False' for each body.
//
// The following values are allowed for collision_behavior, also in any case: none, elastic, subsume, fragment.
// If no value is provided, then 'elastic' is defaulted.
//
// Refer to globals for color values. They can be provided in any case. If not provided, a random color is
// selected for each body in the CSV
//
// Example:
// 100,100,100,100,100,100,10,.5,,,blue
// 1,1,1,1,1,1,10000,10,true,elastic
//
// The above example would load a simulation with one non-sun body, and one sun, both with elastic collision
// behavior.
//
// args:
//  pathSpec                 Path of the CSV
//  bodyCount                Max number of bodies to read from the CSV. To guarantee inclusion of
//                           a body from the CSV, place it earlier in the file than this value.
//  defaultCollisionBehavior The default collision behavior for each body, if not specified in the CSV
//  defaultBodyColor         If not specified in the CSV
//
// returns: the parsed list of bodies
//
func FromCsv(csvPath string, bodyCount int, defaultCollisionBehavior globals.CollisionBehavior,
	defaultBodyColor globals.BodyColor) []*body.Body {

	var bodies []*body.Body
	file, err := os.Open(csvPath)
	defer file.Close()
	if err != nil {
		println("Error opening csv: " + csvPath)
		return nil
	}
	r := csv.NewReader(bufio.NewReader(file))
	r.FieldsPerRecord = -1 // variable
	r.Comment = '#'
	for lines := 0; lines < bodyCount; {
		fields, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		// using a function to parse allows inline panic recovery so we can just skip records with
		// parse errors
		func() {
			defer func() {
				if r := recover(); r != nil {
					// ignore all errors and keep going
				}
			}()
			x := parseFloat(strings.TrimSpace(fields[0]))
			y := parseFloat(strings.TrimSpace(fields[1]))
			z := parseFloat(strings.TrimSpace(fields[2]))
			vx := parseFloat(strings.TrimSpace(fields[3]))
			vy := parseFloat(strings.TrimSpace(fields[4]))
			vz := parseFloat(strings.TrimSpace(fields[5]))
			mass := parseFloat(strings.TrimSpace(fields[6]))
			radius := parseFloat(strings.TrimSpace(fields[7]))
			isSun := false
			if len(fields) >= 9 {
				isSun = parseBool(strings.TrimSpace(fields[8]))
			}
			collisionBehavior := defaultCollisionBehavior
			if len(fields) >= 10 {
				collisionBehavior = globals.ParseCollisionBehavior(strings.TrimSpace(fields[9]))
			}
			bodyColor := defaultBodyColor
			if len(fields) >= 11 {
				bodyColor = globals.ParseBodyColor(strings.TrimSpace(fields[10]))
			}
			fragFactor := float64(0)
			if len(fields) >= 12 {
				fragFactor = parseFloat(strings.TrimSpace(fields[11]))
			}
			fragStep := float64(0)
			if len(fields) >= 13 {
				fragStep = parseFloat(strings.TrimSpace(fields[12]))
			}
			b := body.NewBody(body.NextId(), x, y, z, vx, vy, vz, mass, radius, collisionBehavior, bodyColor, fragFactor,
				fragStep, false, "", "", false)
			if isSun {
				b.SetSun(100)
			}
			bodies = append(bodies, b)
			lines++
		}()
	}
	return bodies
}

//
// parses or panics. The panic is handled by the caller's defer so it's a lean way to parse and just
// skip over errors
//
func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic("ignoreme")
	}
	return f
}

func parseBool(s string) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		panic("ignoreme")
	}
	return b
}
