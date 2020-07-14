package counters

import (
	"sync"
)

// InMemoryCounters uses memory as a values storage
type InMemoryCounters struct {
	values sync.Map
}

// NewInMemoryCounters creates new counter, that uses memory as a storage
func NewInMemoryCounters(initWith map[string]int64) *InMemoryCounters {
	counter := &InMemoryCounters{
		values: sync.Map{},
	}

	for key, value := range initWith {
		counter.values.Store(key, value)
	}

	return counter
}

// Increment increases by 1 counter with passed name
func (c *InMemoryCounters) Increment(name string) {
	if actual, loaded := c.values.LoadOrStore(name, int64(1)); loaded {
		c.values.Store(name, actual.(int64)+1)
	}
}

// Decrement decreases by 1 counter with passed name
func (c *InMemoryCounters) Decrement(name string) {
	if actual, loaded := c.values.LoadOrStore(name, int64(-1)); loaded {
		c.values.Store(name, actual.(int64)-1)
	}
}

// Get returns counter value by name, or 0 in counter was not initialized
func (c *InMemoryCounters) Get(name string) int64 {
	if actual, ok := c.values.Load(name); ok {
		return actual.(int64)
	}

	return 0
}

// Set sets counter value by name
func (c *InMemoryCounters) Set(name string, value int64) {
	c.values.Store(name, value)
}
