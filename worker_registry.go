package main

import (
	"sync"
)

type WorkerRegistry struct {
	sync.RWMutex
	workers map[string]bool
}

func NewWorkerRegistry() *WorkerRegistry {
	return &WorkerRegistry{
		workers: map[string]bool{},
	}
}

func (w *WorkerRegistry) Exists(worker string) bool {
	w.RLock()
	defer w.RUnlock()
	_, ok := w.workers[worker]
	return ok
}

func (w *WorkerRegistry) Add(worker string) {
	w.Lock()
	defer w.Unlock()
	log.Tracef("Adding %s to worker registry", worker)
	w.workers[worker] = true
}

func (w *WorkerRegistry) Remove(worker string) error {
	w.Lock()
	defer w.Unlock()
	log.Tracef("Removing %s from worker registry", worker)
	delete(w.workers, worker)
	return nil
}
