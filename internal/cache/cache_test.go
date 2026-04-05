package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCache_SetAndGet(t *testing.T) {
	c := New[string, int](5 * time.Minute)

	c.Set("key1", 42)
	val, ok := c.Get("key1")

	assert.True(t, ok)
	assert.Equal(t, 42, val)
}

func TestCache_GetMissing(t *testing.T) {
	c := New[string, int](5 * time.Minute)

	val, ok := c.Get("nonexistent")

	assert.False(t, ok)
	assert.Equal(t, 0, val)
}

func TestCache_Expiration(t *testing.T) {
	c := New[string, string](50 * time.Millisecond)

	c.Set("key1", "value1")

	// Should be present immediately
	val, ok := c.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", val)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	val, ok = c.Get("key1")
	assert.False(t, ok)
	assert.Equal(t, "", val)
}

func TestCache_Overwrite(t *testing.T) {
	c := New[string, int](5 * time.Minute)

	c.Set("key1", 1)
	c.Set("key1", 2)

	val, ok := c.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, 2, val)
}

func TestCache_MultipleKeys(t *testing.T) {
	c := New[string, string](5 * time.Minute)

	c.Set("a", "alpha")
	c.Set("b", "beta")
	c.Set("c", "gamma")

	tests := []struct {
		key      string
		expected string
	}{
		{"a", "alpha"},
		{"b", "beta"},
		{"c", "gamma"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			val, ok := c.Get(tt.key)
			assert.True(t, ok)
			assert.Equal(t, tt.expected, val)
		})
	}
}

func TestCache_IntKeys(t *testing.T) {
	c := New[int, string](5 * time.Minute)

	c.Set(1, "one")
	c.Set(2, "two")

	val, ok := c.Get(1)
	assert.True(t, ok)
	assert.Equal(t, "one", val)

	val, ok = c.Get(2)
	assert.True(t, ok)
	assert.Equal(t, "two", val)
}

type structKey struct {
	id   int
	lang string
}

func TestCache_StructKeys(t *testing.T) {
	c := New[structKey, string](5 * time.Minute)

	c.Set(structKey{1, "fr"}, "bonjour")
	c.Set(structKey{1, "en"}, "hello")

	val, ok := c.Get(structKey{1, "fr"})
	assert.True(t, ok)
	assert.Equal(t, "bonjour", val)

	val, ok = c.Get(structKey{1, "en"})
	assert.True(t, ok)
	assert.Equal(t, "hello", val)

	_, ok = c.Get(structKey{2, "fr"})
	assert.False(t, ok)
}
