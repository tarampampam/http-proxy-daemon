package main

type ICounter interface {
	Increment(name string) int64
	Decrement(name string) int64
	Exists(name string) bool
	Get(name string) int64
	Set(name string, value int64)
	All() map[string]int64
}

type Counters struct {
	values map[string]int64
}

// Create new counters instance.
func NewCounters(v map[string]int64) *Counters {
	if v == nil {
		v = make(map[string]int64)
	}
	return &Counters{
		values: v,
	}
}

// Init empty counter by name.
func (c *Counters) initValue(name string) int64 {
	if !c.Exists(name) {
		c.values[name] = 0
	}
	return c.values[name]
}

// Increment counter by name.
func (c *Counters) Increment(name string) int64 {
	c.initValue(name)
	c.values[name]++
	return c.values[name]
}

// Decrement counter by name.
func (c *Counters) Decrement(name string) int64 {
	c.values[name]--
	return c.values[name]
}

// Check counter exists.
func (c *Counters) Exists(name string) bool {
	_, exists := c.values[name]
	return exists
}

// Get current counter state.
func (c *Counters) Get(name string) int64 {
	if c.Exists(name) {
		return c.values[name]
	}
	return 0
}

// Set any counter value.
func (c *Counters) Set(name string, value int64) {
	c.values[name] = value
}

// Get all counters map.
func (c *Counters) All() map[string]int64 {
	return c.values
}
