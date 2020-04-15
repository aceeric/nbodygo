package body

type FragmentationCalcResult struct {
	shouldFragment bool
	thisFactor float32
	otherFactor float32
}

func NoFragmentation() FragmentationCalcResult {
	return FragmentationCalcResult {
		false, 0, 0,
	}
}

func Fragmentation(thisFactor float32, otherFactor float32) FragmentationCalcResult {
	return FragmentationCalcResult {
		true, thisFactor, otherFactor,
	}
}