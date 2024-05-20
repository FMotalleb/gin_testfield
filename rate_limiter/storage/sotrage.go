package rlstorage

type RLStorage interface {
	Get(string) uint16
	Increase(string)
	Decrease(string)
	Free(string)
}
