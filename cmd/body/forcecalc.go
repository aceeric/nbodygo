package body

type ForceCalcResult struct {
	dist float32
	collided bool
}

func NoCollision() ForceCalcResult {
	return ForceCalcResult{0, false}
}

func Collision(dist float32) ForceCalcResult {
	return ForceCalcResult{dist, true}
}