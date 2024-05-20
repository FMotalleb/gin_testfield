package ratelimiter

import (
	"sync"
)

type Storage interface {
	Get(string) uint16
	Increase(string)
	Decrease(string)
	Free(string)
}

type HashMapStorage struct {
	storage map[string]uint16
	lock    sync.Mutex
}

// Decrease implements Storage.
func (h *HashMapStorage) Decrease(id string) {
	defer h.lock.Unlock()
	h.lock.Lock()
	count := h.storage[id]
	if count <= 1 {
		h.Free(id)
	} else {
		h.storage[id] = count - 1
	}
}

func (h *HashMapStorage) Free(id string) {
	defer h.lock.Unlock()
	h.lock.Lock()
	delete(h.storage, id)
}

// Get implements Storage.
func (h *HashMapStorage) Get(id string) uint16 {
	return h.storage[id]
}

// Increase implements Storage.
func (h *HashMapStorage) Increase(id string) {
	defer h.lock.Unlock()
	h.lock.Lock()
	h.storage[id]++
}

func NewHashMapStorage() Storage {
	return &HashMapStorage{
		storage: make(map[string]uint16),
		lock:    sync.Mutex{},
	}
}
