package body

type FragmentationCalcResult struct {
	shouldFragment bool
	thisFactor float64
	otherFactor float64
}

func NoFragmentation() FragmentationCalcResult {
	return FragmentationCalcResult {
		false, 0, 0,
	}
}

func Fragmentation(thisFactor float64, otherFactor float64) FragmentationCalcResult {
	return FragmentationCalcResult {
		true, thisFactor, otherFactor,
	}
}