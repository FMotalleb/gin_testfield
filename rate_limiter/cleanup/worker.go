// cleanupworker.go
package cleanup

import (
	"time"

	rlstorage "github.com/FMotalleb/gin_testfield/rate_limiter/storage"
)

type CleanupWorker struct {
	storage  rlstorage.RLStorage
	rotation time.Duration
	stopChan chan struct{}
}

func NewWorker(storage rlstorage.RLStorage, rotation time.Duration) *CleanupWorker {
	return &CleanupWorker{
		storage:  storage,
		rotation: rotation,
		stopChan: make(chan struct{}),
	}
}

func (cw *CleanupWorker) Start() {
	go cw.run()
}

func (cw *CleanupWorker) Stop() {
	close(cw.stopChan)
}

func (cw *CleanupWorker) run() {
	ticker := time.NewTicker(cw.rotation)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cw.storage.FreeAll()
		case <-cw.stopChan:
			return
		}
	}
}
