package rlstorage

import (
	"sync"
)

// hashMapStorage is a struct that represents a storage implementation using a hash map.
type hashMapStorage struct {
	storage map[string]uint16 // The underlying hash map to store the key-value pairs
	lock    sync.Mutex        // A mutex lock to ensure thread-safe access to the storage
}

// Decrease decrements the count for the given id in the storage.
// If the count becomes 0 or less, the id is removed from the storage.
func (h *hashMapStorage) Decrease(id string) {
	defer h.lock.Unlock()  // Unlock the mutex when the function returns
	h.lock.Lock()          // Lock the mutex to ensure exclusive access to the storage
	count := h.storage[id] // Get the current count for the id
	if count <= 1 {
		h.Free(id) // If the count is 1 or less, remove the id from the storage
	} else {
		h.storage[id] = count - 1 // Otherwise, decrement the count by 1
	}
}

// Free removes the given id from the storage.
func (h *hashMapStorage) Free(id string) {
	defer h.lock.Unlock() // Unlock the mutex when the function returns
	h.lock.Lock()         // Lock the mutex to ensure exclusive access to the storage
	delete(h.storage, id) // Remove the id from the storage
}

// Get retrieves the count for the given id from the storage.
func (h *hashMapStorage) Get(id string) uint16 {
	return h.storage[id] // Return the count for the id (returns 0 if id doesn't exist)
}

// Increase increments the count for the given id in the storage.
func (h *hashMapStorage) Increase(id string) {
	defer h.lock.Unlock() // Unlock the mutex when the function returns
	h.lock.Lock()         // Lock the mutex to ensure exclusive access to the storage
	h.storage[id]++       // Increment the count for the id by 1
}

// NewHashMapStorage creates a new instance of RLStorage using hashMapStorage.
func NewHashMapStorage() RLStorage {
	return &hashMapStorage{
		storage: make(map[string]uint16), // Initialize the hash map storage
		lock:    sync.Mutex{},            // Initialize the mutex lock
	}
}
