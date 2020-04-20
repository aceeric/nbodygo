package body

type ForceCalcResult struct {
	dist float64
	collided bool
}

func NoCollision() ForceCalcResult {
	return ForceCalcResult{0, false}
}

func Collision(dist float64) ForceCalcResult {
	return ForceCalcResult{dist, true}
}