package rlstorage

// RLStorage is an interface that defines the contract for a rate limiting storage mechanism.
// It provides methods for retrieving, incrementing, decrementing, and resetting rate limiting values.
type RLStorage interface {
	// Get retrieves the current rate value associated with the given ID.
	// It returns the value as a uint16 (an unsigned 16-bit integer).
	Get(string) uint16

	// Increase increments the rate value associated with the given ID.
	Increase(string)

	// Decrease decrements the rate value associated with the given ID.
	Decrease(string)

	// Free resets or frees the rate value associated with the given ID,
	// typically by setting it to zero or removing it from storage.
	Free(string)

	// Free resets or frees the rate value of all IDs
	FreeAll()
}
