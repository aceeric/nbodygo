package body

type forceCalcResult struct {
	dist float64
	collided bool
}

func noCollision() forceCalcResult {
	return forceCalcResult{0, false}
}

func collision(dist float64) forceCalcResult {
	return forceCalcResult{dist, true}
}