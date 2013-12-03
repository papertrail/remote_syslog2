package main

import (
	"container/list"
	"fmt"
	"sync"
)

func NewWorkerRegistry() WorkerRegistry {
	return WorkerRegistry{workers: list.New()}
}

type WorkerRegistry struct {
	workers *list.List
	mu      sync.RWMutex
}

func (w *WorkerRegistry) Exists(worker string) bool {
	_, err := w.Find(worker)
	if err != nil {
		return false
	}
	return true
}

func (w *WorkerRegistry) Find(worker string) (*list.Element, error) {
	defer w.mu.RUnlock()
	w.mu.RLock()
	log.Tracef("Called thread safe Find")
	log.Tracef("Looking for %s in the worker registry", worker)
	element, err := w.find(worker)
	if err != nil {
		log.Tracef("%s was not found in the worker registry", worker)
		return nil, err
	}
	log.Tracef("%s was found in the worker registry", worker)
	return element, nil
}

func (w *WorkerRegistry) find(worker string) (*list.Element, error) {
	log.Tracef("Called non-thread safe find")
	for e := w.workers.Front(); e != nil; e = e.Next() {
		log.Tracef("Checking if %s == %s", worker, e.Value)
		if e.Value == worker {
			return e, nil
		}
	}
	return nil, fmt.Errorf("Failed to find worker %s", worker)
}

func (w *WorkerRegistry) Add(worker string) {
	defer w.mu.Unlock()
	w.mu.Lock()
	log.Tracef("Adding %s to worker registry", worker)
	w.workers.PushBack(worker)
}

func (w *WorkerRegistry) Remove(worker string) error {
	defer w.mu.Unlock()
	w.mu.Lock()
	log.Tracef("Removing %s from worker registry", worker)
	workerElement, err := w.find(worker)
	if err != nil {
		log.Tracef("Failed to remove worker: %s", err)
		return err
	}
	w.workers.Remove(workerElement)
	log.Tracef("Removed %s from worker registry", worker)
	return nil
}
