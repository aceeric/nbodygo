package globals

import (
	"nbodygo/internal/pkg/math32"
	"strconv"
	"strings"
)

type CollisionBehavior int

const (
	None     CollisionBehavior = 0
	Subsume  CollisionBehavior = 1
	Elastic  CollisionBehavior = 2
	Fragment CollisionBehavior = 3
)

type BodyColor int

// TODO ENUM
const (
	Random    BodyColor = 0
	Black     BodyColor = 1
	White     BodyColor = 2
	Darkgray  BodyColor = 3
	Gray      BodyColor = 4
	Lightgray BodyColor = 5
	Red       BodyColor = 6
	Green     BodyColor = 7
	Blue      BodyColor = 8
	Yellow    BodyColor = 9
	Magenta   BodyColor = 10
	Cyan      BodyColor = 11
	Orange    BodyColor = 12
	Brown     BodyColor = 13
	Pink      BodyColor = 14
)

func ParseCollisionBehavior(s string) CollisionBehavior {
	for i, item := range []string{"none", "subsume", "elastic", "fragment"} {
		if item == strings.ToLower(s) {
			return CollisionBehavior(i)
		}
	}
	return Elastic
}

func parseBoolean(s string) bool {
	switch strings.ToLower(s) {
	case "t", "true", "1", "y", "yes":
		return true
	}
	return false
}

func ParseBodyColor(s string) BodyColor {
	for i, item := range []string{"random", "black", "white", "darkgray", "gray", "lightgray", "red",
		"green", "blue", "yellow", "magenta", "cyan", "orange", "brown", "pink"} {
		if item == strings.ToLower(s) {
			return BodyColor(i)
		}
	}
	return Random
}

func ParseVector(s string) *math32.Vector3 {
	components := strings.Split(s, ",")
	if len(components) != 3 {
		return &math32.Vector3{}
	}
	x, _ := strconv.ParseFloat(components[0], 32)
	y, _ := strconv.ParseFloat(components[1], 32)
	z, _ := strconv.ParseFloat(components[2], 32)
	return &math32.Vector3{
		X: float32(x), Y: float32(y), Z: float32(z),
	}
}
