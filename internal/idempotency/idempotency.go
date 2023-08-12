package idempotency

import (
	"sync"

	"k8s.io/klog/v2"
)

type Idempotency struct {
	lock *sync.Mutex

	// Reccord the volume that is being handled
	processing map[string]bool
}

func NewIdempotency() *Idempotency {
	return &Idempotency{
		lock:       &sync.Mutex{},
		processing: make(map[string]bool),
	}
}

func (i *Idempotency) IsProcessing(volumeId string) bool {
	i.lock.Lock()
	defer i.lock.Unlock()

	_, ok := i.processing[volumeId]
	return ok
}

// Caller should make sure there is no volume is being handled(call isProcessing() first)
func (i *Idempotency) AddProcessing(volumeId string) {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.processing[volumeId] = true
}

func (i *Idempotency) RemoveProcessing(volumeId string) {
	i.lock.Lock()
	defer i.lock.Unlock()

	klog.V(4).InfoS("Remove volume from processing list", "volumeId", volumeId)
	delete(i.processing, volumeId)
}
