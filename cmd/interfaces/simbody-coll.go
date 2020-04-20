package interfaces

type SimBodyCollection interface {
	Add(SimBody)
	IterateOnce(callback func(item SimBody))
	GetArrayCopy() *[]SimBody
	Cycle()
}
