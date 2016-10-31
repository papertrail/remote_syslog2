package main

import (
	"sync"
)

// WorkerRegistry keeps track of which files are being actively tailed in order to avoid sending the same logs
// multiple times. Implementations must be thread-safe.
type WorkerRegistry interface {
	// Exists returns true if a log file is currently being tailed
	Exists(worker string) bool

	// Add marks a log file as being currently tailed
	Add(worker string)

	// Remove clears a log file from the registry
	Remove(worker string)
}

// InMemoryRegistry is a simple WorkerRegistry implementation that uses a map protected by a sync.RWMutex.
type InMemoryRegistry struct {
	mu      sync.RWMutex
	workers map[string]bool
}

func NewInMemoryRegistry() WorkerRegistry {
	return &InMemoryRegistry{workers: make(map[string]bool)}
}

func (imr *InMemoryRegistry) Exists(worker string) bool {
	imr.mu.RLock()
	defer imr.mu.RUnlock()
	_, ok := imr.workers[worker]
	return ok
}

func (imr *InMemoryRegistry) Add(worker string) {
	imr.mu.Lock()
	defer imr.mu.Unlock()
	log.Tracef("Adding %s to worker registry", worker)
	imr.workers[worker] = true
}

func (imr *InMemoryRegistry) Remove(worker string) {
	imr.mu.Lock()
	defer imr.mu.Unlock()
	log.Tracef("Removing %s from worker registry", worker)
	delete(imr.workers, worker)
}
