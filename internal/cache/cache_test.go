package cache

import (
	"bytes"
	"testing"
	"time"
)

func TestCache_Add(t *testing.T) {
	// Create a new cache instance
	cache := New(time.Minute)

	// Add a cache element with a key and data
	key := "test"
	data := []byte("test data")
	cache.Add(key, data)

	// Check if the cache element exists in the store
	if _, ok := cache.store[key]; !ok {
		t.Errorf("Cache element with key '%s' not found", key)
	}

	// Check if the data of the cache element is correct
	if got := cache.store[key].data; !bytes.Equal(got, data) {
		t.Errorf("Cache element data mismatch, got %v, want %v", got, data)
	}
}

func TestCache_Get(t *testing.T) {
	// Create a new cache instance
	cache := New(time.Minute)

	// Add a cache element with a key and data
	key := "test"
	data := []byte("test data")
	cache.Add(key, data)

	// Retrieve the cache element using the Get method
	got, ok := cache.Get(key)

	// Check if the cache element is found
	if !ok {
		t.Errorf("Cache element with key '%s' not found", key)
	}

	// Check if the retrieved data is correct
	if string(got) != string(data) {
		t.Errorf("Retrieved data mismatch, got %s, want %s", got, data)
	}
}

func TestCache_Get_Expired(t *testing.T) {
	// Create a new cache instance with a short timeout duration
	cache := New(1 * time.Millisecond)

	// Add a cache element with a key and data
	key := "test"
	data := []byte("test data")
	cache.Add(key, data)

	// Wait for the cache element to expire
	time.Sleep(2 * time.Millisecond)

	// Retrieve the cache element using the Get method
	got, ok := cache.Get(key)

	// Check if the cache element is not found
	if ok {
		t.Errorf("Cache element with key '%s' found, expected not found", key)
	}

	// Check if the retrieved data is nil
	if got != nil {
		t.Errorf("Retrieved data mismatch, got %v, want nil", got)
	}
}

func TestCache_Purge(t *testing.T) {
	// Create a new cache instance
	cache := New(time.Minute)

	// Add multiple cache elements
	cache.Add("key1", []byte("data1"))
	cache.Add("key2", []byte("data2"))

	// Purge the cache
	cache.Purge()

	// Check if the store is empty
	if len(cache.store) != 0 {
		t.Errorf("Cache store is not empty after purging")
	}
}
