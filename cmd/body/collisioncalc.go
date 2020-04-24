package body

type collisionCalcResult struct {
	collided bool
	vx1, vy1, vz1 float64
	vx2, vy2, vz2 float64
	vx_cm, vy_cm, vz_cm float64
}

func elasticCollision(vx1, vy1, vz1, vx2, vy2, vz2, vx_cm, vy_cm, vz_cm float64) collisionCalcResult {
	return collisionCalcResult{
		collided: true,
		vx1: vx1, vy1: vy1, vz1: vz1,
		vx2: vx2, vy2: vy2, vz2: vz2,
		vx_cm: vx_cm, vy_cm: vy_cm, vz_cm: vz_cm,
	}
}

func noElasticCollision() collisionCalcResult {
	return collisionCalcResult{
		collided: false,
		vx1: 0, vy1: 0, vz1: 0,
		vx2: 0, vy2: 0, vz2: 0,
		vx_cm: 0, vy_cm: 0, vz_cm: 0,
	}
}
