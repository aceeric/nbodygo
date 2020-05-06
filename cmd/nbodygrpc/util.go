package nbodygrpc

import "nbodygo/cmd/globals"

func GrpcColorToSimColor(color BodyColorEnum) globals.BodyColor {
	switch color {
	case BodyColorEnum_RANDOM:
		return globals.Random
	case BodyColorEnum_BLACK:
		return globals.Black
	case BodyColorEnum_WHITE:
		return globals.White
	case BodyColorEnum_DARKGRAY:
		return globals.Darkgray
	case BodyColorEnum_GRAY:
		return globals.Gray
	case BodyColorEnum_LIGHTGRAY:
		return globals.Lightgray
	case BodyColorEnum_RED:
		return globals.Red
	case BodyColorEnum_GREEN:
		return globals.Green
	case BodyColorEnum_BLUE:
		return globals.Blue
	case BodyColorEnum_YELLOW:
		return globals.Yellow
	case BodyColorEnum_MAGENTA:
		return globals.Magenta
	case BodyColorEnum_CYAN:
		return globals.Cyan
	case BodyColorEnum_ORANGE:
		return globals.Orange
	case BodyColorEnum_BROWN:
		return globals.Brown
	case BodyColorEnum_PINK:
		return globals.Pink
	case BodyColorEnum_NOCOLOR:
		fallthrough
	default:
		return globals.Random
	}
}

func SimColorToGrpcColor(color globals.BodyColor) BodyColorEnum {
	switch color {
	case globals.Black:
		return BodyColorEnum_BLACK
	case globals.White:
		return BodyColorEnum_WHITE
	case globals.Darkgray:
		return BodyColorEnum_DARKGRAY
	case globals.Gray:
		return BodyColorEnum_GRAY
	case globals.Lightgray:
		return BodyColorEnum_LIGHTGRAY
	case globals.Red:
		return BodyColorEnum_RED
	case globals.Green:
		return BodyColorEnum_GREEN
	case globals.Blue:
		return BodyColorEnum_BLUE
	case globals.Yellow:
		return BodyColorEnum_YELLOW
	case globals.Magenta:
		return BodyColorEnum_MAGENTA
	case globals.Cyan:
		return BodyColorEnum_CYAN
	case globals.Orange:
		return BodyColorEnum_ORANGE
	case globals.Brown:
		return BodyColorEnum_BROWN
	case globals.Pink:
		return BodyColorEnum_PINK
	case globals.Random:
		fallthrough
	default:
		return BodyColorEnum_RANDOM
	}
}

func GrpcCbToSimCb(behavior CollisionBehaviorEnum) globals.CollisionBehavior {
	switch behavior {
	case CollisionBehaviorEnum_NONE:
		return globals.None
	case CollisionBehaviorEnum_SUBSUME:
		return globals.Subsume
	case CollisionBehaviorEnum_FRAGMENT:
		return globals.Fragment
	case CollisionBehaviorEnum_ELASTIC:
		fallthrough
	default:
		return globals.Elastic
	}
}

func SimCbToGrpcCb(behavior globals.CollisionBehavior) CollisionBehaviorEnum {
	switch behavior {
	case globals.None:
		return CollisionBehaviorEnum_NONE
	case globals.Subsume:
		return CollisionBehaviorEnum_SUBSUME
	case globals.Fragment:
		return CollisionBehaviorEnum_FRAGMENT
	default:
		return CollisionBehaviorEnum_ELASTIC
	}
}
