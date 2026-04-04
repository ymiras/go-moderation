package storage

import (
	"container/list"
	"sync"
)

// Cache is a generic LRU cache.
type Cache[K comparable, V any] struct {
	maxEntries int
	onEvict    func(K, V)
	mu         sync.Mutex
	items      map[K]*list.Element
	order      *list.List
}

// entry is the internal type stored in the linked list.
type entry[K comparable, V any] struct {
	key   K
	value V
}

// NewCache creates a new LRU cache with the specified maximum entries.
func NewCache[K comparable, V any](maxEntries int, onEvict func(K, V)) *Cache[K, V] {
	return &Cache[K, V]{
		maxEntries: maxEntries,
		onEvict:    onEvict,
		items:      make(map[K]*list.Element),
		order:      list.New(),
	}
}

// Set adds or updates an entry in the cache.
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.items[key]; exists {
		// Update existing entry - move to front and update value
		c.order.MoveToFront(elem)
		elem.Value.(*entry[K, V]).value = value
		return
	}

	// Add new entry
	elem := c.order.PushFront(&entry[K, V]{key: key, value: value})
	c.items[key] = elem

	// Evict if over capacity
	if c.maxEntries > 0 && c.order.Len() > c.maxEntries {
		c.evictOldest()
	}
}

// Get retrieves an entry from the cache.
// Returns the value and true if found, or zero value and false if not found.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, exists := c.items[key]
	if !exists {
		var zero V
		return zero, false
	}

	// Move to front (most recently used)
	c.order.MoveToFront(elem)
	return elem.Value.(*entry[K, V]).value, true
}

// Delete removes an entry from the cache.
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, exists := c.items[key]
	if !exists {
		return
	}

	c.order.Remove(elem)
	delete(c.items, key)
}

// Len returns the number of entries in the cache.
func (c *Cache[K, V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.order.Len()
}

// Clear removes all entries from the cache.
func (c *Cache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.onEvict != nil {
		for _, elem := range c.items {
			e := elem.Value.(*entry[K, V])
			c.onEvict(e.key, e.value)
		}
	}

	c.items = make(map[K]*list.Element)
	c.order.Init()
}

// evictOldest removes the least recently used entry.
func (c *Cache[K, V]) evictOldest() {
	elem := c.order.Back()
	if elem == nil {
		return
	}

	c.order.Remove(elem)
	e := elem.Value.(*entry[K, V])
	delete(c.items, e.key)

	if c.onEvict != nil {
		c.onEvict(e.key, e.value)
	}
}
