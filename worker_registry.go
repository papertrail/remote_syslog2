package main

import (
	"sync"
)

func NewWorkerRegistry() WorkerRegistry {
	return WorkerRegistry{workers: make(map[string]bool)}
}

type WorkerRegistry struct {
	workers map[string]bool
	mu      sync.RWMutex
}

func (w *WorkerRegistry) Exists(worker string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	_, ok := w.workers[worker]
	return ok
}

func (w *WorkerRegistry) Add(worker string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	log.Tracef("Adding %s to worker registry", worker)
	w.workers[worker] = true
}

func (w *WorkerRegistry) Remove(worker string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	log.Tracef("Removing %s from worker registry", worker)
	delete(w.workers, worker)
	return nil
}
