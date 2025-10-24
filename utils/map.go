package utils

import (
	"sync"
)

type Map[K comparable, V any] struct {
	data *sync.Map
	lock *sync.RWMutex
}

func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{
		data: new(sync.Map),
		lock: new(sync.RWMutex),
	}
}

func (g *Map[K, V]) Clear() {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.data = new(sync.Map) // 简单实现，替换为新的sync.Map
}

func (g *Map[K, V]) Store(key K, value V) {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.data.Store(key, value)
}

func (g *Map[K, V]) Load(key K) (value V, ok bool) {
	g.lock.RLock()
	defer g.lock.RUnlock()
	val, ok := g.data.Load(key)
	if !ok {
		return value, false
	}
	return val.(V), true
}

func (g *Map[K, V]) LoadOrStore(key K, value V) (V, bool) {
	g.lock.Lock()
	defer g.lock.Unlock()
	loadedVal, loaded := g.data.LoadOrStore(key, value)
	return loadedVal.(V), loaded
}

func (g *Map[K, V]) LoadAndDelete(key K) (value V, ok bool) {
	g.lock.Lock()
	defer g.lock.Unlock()
	val, loaded := g.data.LoadAndDelete(key)
	if !loaded {
		return value, false
	}
	return val.(V), true
}

func (g *Map[K, V]) RemoveIf(key K, f func(V) bool) bool {
	g.lock.Lock()
	defer g.lock.Unlock()
	val, ok := g.data.Load(key)
	if !ok {
		return false
	}
	if f(val.(V)) {
		g.data.Delete(key)
		return true
	}
	return false
}

func (g *Map[K, V]) Range(f func(key K, value V) bool) {
	g.lock.RLock()
	defer g.lock.RUnlock()
	g.data.Range(func(key, value interface{}) bool {
		return f(key.(K), value.(V))
	})
}

func (g *Map[K, V]) Len() int {
	g.lock.RLock()
	defer g.lock.RUnlock()

	l := 0
	// 遍历所有 key，主动检查是否存在（减少弱一致性影响）
	g.data.Range(func(key, value interface{}) bool {
		_, exists := g.data.Load(key)
		if exists {
			l++
		}
		return true
	})
	return l
}
