package storage

import (
	"testing"
)

func TestNewCache(t *testing.T) {
	c := NewCache[string, string](10, nil)
	if c == nil {
		t.Fatal("NewCache() returned nil")
	}
}

func TestCacheSetAndGet(t *testing.T) {
	c := NewCache[string, int](10, nil)

	c.Set("k1", 100)
	v, ok := c.Get("k1")
	if !ok {
		t.Error("Get(k1) = false, want true")
	}
	if v != 100 {
		t.Errorf("Get(k1) = %d, want 100", v)
	}
}

func TestCacheGetMiss(t *testing.T) {
	c := NewCache[string, string](10, nil)

	v, ok := c.Get("nonexistent")
	if ok {
		t.Error("Get(nonexistent) = true, want false")
	}
	if v != "" {
		t.Errorf("Get(nonexistent) = %q, want empty", v)
	}
}

func TestCacheLRUEviction(t *testing.T) {
	c := NewCache[string, int](3, nil)

	c.Set("k1", 1)
	c.Set("k2", 2)
	c.Set("k3", 3)

	// k1, k2, k3 in that order (k1 oldest)
	// Adding k4 should evict k1

	c.Set("k4", 4)

	_, ok := c.Get("k1")
	if ok {
		t.Error("k1 should have been evicted")
	}

	// k2 and k3 should still exist
	v, ok := c.Get("k2")
	if !ok || v != 2 {
		t.Error("k2 should still exist")
	}

	v, ok = c.Get("k3")
	if !ok || v != 3 {
		t.Error("k3 should still exist")
	}

	v, ok = c.Get("k4")
	if !ok || v != 4 {
		t.Error("k4 should exist")
	}
}

func TestCacheUpdate(t *testing.T) {
	c := NewCache[string, int](10, nil)

	c.Set("k1", 100)
	c.Set("k1", 200) // update

	v, ok := c.Get("k1")
	if !ok {
		t.Error("Get(k1) = false, want true")
	}
	if v != 200 {
		t.Errorf("Get(k1) = %d, want 200", v)
	}

	if c.Len() != 1 {
		t.Errorf("Len() = %d, want 1", c.Len())
	}
}

func TestCacheDelete(t *testing.T) {
	c := NewCache[string, int](10, nil)

	c.Set("k1", 100)
	c.Delete("k1")

	_, ok := c.Get("k1")
	if ok {
		t.Error("k1 should have been deleted")
	}

	if c.Len() != 0 {
		t.Errorf("Len() = %d, want 0", c.Len())
	}
}

func TestCacheClear(t *testing.T) {
	c := NewCache[string, int](10, nil)

	c.Set("k1", 1)
	c.Set("k2", 2)
	c.Clear()

	if c.Len() != 0 {
		t.Errorf("Len() = %d, want 0", c.Len())
	}
}

func TestCacheOnEvictCallback(t *testing.T) {
	evicted := make(map[string]int)

	c := NewCache[string, int](2, func(k string, v int) {
		evicted[k] = v
	})

	c.Set("k1", 1)
	c.Set("k2", 2)
	c.Set("k3", 3) // should evict k1

	if evicted["k1"] != 1 {
		t.Errorf("evicted[k1] = %d, want 1", evicted["k1"])
	}

	// k2 should not have been evicted
	if _, ok := evicted["k2"]; ok {
		t.Error("k2 should not have been evicted")
	}
}

func TestCacheUnlimitedCapacity(t *testing.T) {
	c := NewCache[string, int](0, nil) // 0 = unlimited

	for i := 0; i < 100; i++ {
		c.Set(string(rune('a'+i)), i)
	}

	if c.Len() != 100 {
		t.Errorf("Len() = %d, want 100", c.Len())
	}
}

func TestCacheConcurrentAccess(t *testing.T) {
	c := NewCache[string, int](100, nil)

	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 50; i++ {
			c.Set("key", i)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 50; i++ {
			c.Get("key")
		}
		done <- true
	}()

	<-done
	<-done
}
