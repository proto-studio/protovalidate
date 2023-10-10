package objects

import (
	"fmt"
	"sync"
)

// counter is used for evaluating rules and keeps track of how many rules
// are remaining for each key.
type counter struct {
	mu    sync.RWMutex // mu protects concurrent access to count.
	count int          // count holds the current value of the counter.
	cond  *sync.Cond   // cond is used to signal when the counter reaches 0.
	name  string       // name is used primarily for debugging.
}

// newCounter initializes and returns a new counter object.
func newCounter(name string) *counter {
	c := &counter{
		name: name,
	}
	c.cond = sync.NewCond(&c.mu)
	return c
}

// Increment safely increases the counter by 1.
func (c *counter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++
}

// Clears the counter immediately.
func (c *counter) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.count > 0 {
		c.count = 0
		c.cond.Broadcast()
	}
}

// Lock locks the counter for writing.
func (c *counter) Lock() {
	c.mu.Lock()
}

// Unlock decreases the counter by 1 and then unlocks it.
// If the counter reaches 0, it broadcasts to any waiting goroutines.
func (c *counter) Unlock() {
	c.count--
	count := c.count
	c.mu.Unlock()

	if count == 0 {
		c.cond.Broadcast()
	} else if count < 0 {
		panic(fmt.Errorf("negative rule counter: %d", count))
	}

}

// Wait waits for the counter to reach 0.
func (c *counter) Wait() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for c.count > 0 {
		c.cond.Wait()
	}
}

// counterSet manages a thread-safe collection of counters, each associated with a unique key.
type counterSet struct {
	mu       sync.RWMutex        // mu protects concurrent access to counters.
	counters map[string]*counter // counters holds the collection of counters.
}

// newCounterSet initializes and returns a new counterSet object.
func newCounterSet() *counterSet {
	return &counterSet{
		counters: make(map[string]*counter),
	}
}

// Increment safely increases the counter associated with the given key by 1.
// If a counter doesn't exist for the key, it creates one.
func (cs *counterSet) Increment(key string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if _, exists := cs.counters[key]; !exists {
		cs.counters[key] = newCounter(key)
	}
	cs.counters[key].Increment()
}

// Clear safely sets the counter to zero.
// Used when we know for a fact that no future rules will succeed for that key.
func (cs *counterSet) Clear(key string) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	if _, exists := cs.counters[key]; exists {
		cs.counters[key].Clear()
	}
}

// Lock locks the counter for a specific key for writing.
func (cs *counterSet) Lock(key string) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	if _, exists := cs.counters[key]; exists {
		cs.counters[key].Lock()
	}
}

// Unlock unlocks the counter for a specific key for writing.
func (cs *counterSet) Unlock(key string) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	if _, exists := cs.counters[key]; exists {
		cs.counters[key].Unlock()
	}
}

// Wait waits for the counters associated with the provided keys to reach 0.
// If a key doesn't have an associated counter, it simply moves on to the next key.
func (cs *counterSet) Wait(keys ...string) {
	for _, key := range keys {
		cs.mu.RLock()
		c, exists := cs.counters[key]
		cs.mu.RUnlock()

		if exists {
			c.Wait()
		}
	}
}
