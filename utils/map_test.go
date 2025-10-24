package utils

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMap_BasicOps(t *testing.T) {
	m := NewMap[string, int]()

	// Store & Load
	m.Store("a", 1)
	val, ok := m.Load("a")
	assert.True(t, ok)
	assert.Equal(t, 1, val)

	// Load non-existent key
	_, ok = m.Load("b")
	assert.False(t, ok)

	// LoadOrStore
	loadedVal, loaded := m.LoadOrStore("a", 2)
	assert.True(t, loaded)
	assert.Equal(t, 1, loadedVal)

	loadedVal, loaded = m.LoadOrStore("b", 2)
	assert.False(t, loaded)
	assert.Equal(t, 2, loadedVal)

	// LoadAndDelete
	delVal, ok := m.LoadAndDelete("a")
	assert.True(t, ok)
	assert.Equal(t, 1, delVal)
	_, ok = m.Load("a")
	assert.False(t, ok)

	// Range
	sum := 0
	m.Range(func(key string, val int) bool {
		sum += val
		return true
	})
	assert.Equal(t, 2, sum)

	// Len
	assert.Equal(t, 1, m.Len())

	// RemoveIf
	ok = m.RemoveIf("b", func(val int) bool {
		return val == 2
	})
	assert.True(t, ok)
	_, ok = m.Load("b")
	assert.False(t, ok)

	// Clear
	m.Store("c", 3)
	m.Clear()
	assert.Equal(t, 0, m.Len())
}

func TestMap_ConcurrentSafety(t *testing.T) {
	m := NewMap[int, int]()
	var wg sync.WaitGroup
	const goroutines = 100
	const opsPerGoroutine = 1000

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(gID int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				key := gID*opsPerGoroutine + j
				m.Store(key, key)
				// 验证存储的值
				val, ok := m.Load(key)
				assert.True(t, ok)
				assert.Equal(t, key, val)
			}
		}(i)
	}

	wg.Wait()
	assert.Equal(t, goroutines*opsPerGoroutine, m.Len())
}
