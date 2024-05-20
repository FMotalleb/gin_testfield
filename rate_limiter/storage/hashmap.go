package rlstorage

import (
	"sync"

	"github.com/sirupsen/logrus"
)

// hashMapStorage is a struct that represents a storage implementation using a hash map.
type hashMapStorage struct {
	storage map[string]uint16 // The underlying hash map to store the key-value pairs
	lock    sync.Mutex        // A mutex lock to ensure thread-safe access to the storage
	logger  *logrus.Logger    // Logger instance for logging messages
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
	h.logger.Debugf("Freed ID '%s' from storage", id)
}

// Get retrieves the count for the given id from the storage.
func (h *hashMapStorage) Get(id string) uint16 {
	defer h.lock.Unlock() // Unlock the mutex when the function returns
	h.lock.Lock()         // Lock the mutex to ensure exclusive access to the storage
	count := h.storage[id]
	h.logger.Debugf("Got count %d for ID '%s'", count, id)
	return count // Return the count for the id (returns 0 if id doesn't exist)
}

// Increase increments the count for the given id in the storage.
func (h *hashMapStorage) Increase(id string) {
	defer h.lock.Unlock() // Unlock the mutex when the function returns
	h.lock.Lock()         // Lock the mutex to ensure exclusive access to the storage
	h.storage[id]++       // Increment the count for the id by 1
	h.logger.Debugf("Increased count to %d for ID '%s'", h.storage[id], id)
}

// NewHashMapStorage creates a new instance of RLStorage using hashMapStorage.
func NewHashMapStorage(logger *logrus.Logger) RLStorage {
	return &hashMapStorage{
		storage: make(map[string]uint16), // Initialize the hash map storage
		lock:    sync.Mutex{},            // Initialize the mutex lock
		logger:  logger,                  // Set the logger instance
	}
}

// FreeAll removes all entries from the storage.
func (h *hashMapStorage) FreeAll() {
	defer h.lock.Unlock() // Unlock the mutex when the function returns
	h.lock.Lock()         // Lock the mutex to ensure exclusive access to the storage
	h.storage = make(map[string]uint16)
	h.logger.Info("Freed all entries from storage")
}
