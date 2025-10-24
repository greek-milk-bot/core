package utils

import (
	"errors"
	"sync"
)

// Array 是一个线程安全的动态数组实现，元素类型必须是可比较的
type Array[T comparable] struct {
	mu   sync.RWMutex
	data []T
}

// NewArray 创建一个新的线程安全数组
func NewArray[T comparable]() *Array[T] {
	return &Array[T]{
		data: make([]T, 0),
	}
}

// NewArrayWithCapacity 创建一个指定初始容量的线程安全数组
func NewArrayWithCapacity[T comparable](capacity int) *Array[T] {
	if capacity < 0 {
		capacity = 0
	}
	return &Array[T]{
		data: make([]T, 0, capacity),
	}
}

// Len 返回数组的长度
func (a *Array[T]) Len() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.data)
}

// Cap 返回数组的容量
func (a *Array[T]) Cap() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return cap(a.data)
}

// Get 获取指定索引的元素
func (a *Array[T]) Get(index int) (T, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var zero T
	if index < 0 || index >= len(a.data) {
		return zero, errors.New("index out of range")
	}
	return a.data[index], nil
}

// Set 设置指定索引的元素
func (a *Array[T]) Set(index int, value T) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if index < 0 || index >= len(a.data) {
		return errors.New("index out of range")
	}
	a.data[index] = value
	return nil
}

// Append 在数组末尾添加一个或多个元素
func (a *Array[T]) Append(values ...T) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.data = append(a.data, values...)
}

// AddIfNotExists 当元素不存在时添加到数组末尾，返回是否添加成功
func (a *Array[T]) AddIfNotExists(value T) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 检查元素是否已存在
	for _, v := range a.data {
		if v == value {
			return false // 元素已存在，不添加
		}
	}

	// 元素不存在，添加到数组
	a.data = append(a.data, value)
	return true
}

// Insert 在指定索引位置插入元素
func (a *Array[T]) Insert(index int, value T) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if index < 0 || index > len(a.data) {
		return errors.New("index out of range")
	}

	// 分步创建新切片，避免嵌套append导致的数据覆盖
	newData := make([]T, 0, len(a.data)+1)
	newData = append(newData, a.data[:index]...)
	newData = append(newData, value)
	newData = append(newData, a.data[index:]...)
	a.data = newData
	return nil
}

// Delete 删除指定索引的元素
func (a *Array[T]) Delete(index int) (T, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var zero T
	if index < 0 || index >= len(a.data) {
		return zero, errors.New("index out of range")
	}

	// 保存要删除的元素
	value := a.data[index]

	// 删除元素
	a.data = append(a.data[:index], a.data[index+1:]...)
	return value, nil
}

// DeleteByValue 删除第一个匹配的元素，返回是否删除成功
func (a *Array[T]) DeleteByValue(value T) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 查找元素索引
	for i, v := range a.data {
		if v == value {
			// 删除找到的元素
			a.data = append(a.data[:i], a.data[i+1:]...)
			return true
		}
	}

	// 未找到元素
	return false
}

// Clear 清空数组
func (a *Array[T]) Clear() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.data = a.data[:0] // 保留容量，清空元素
}

// Slice 返回数组的副本切片
func (a *Array[T]) Slice() []T {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// 返回副本以防止外部修改内部数据
	result := make([]T, len(a.data))
	copy(result, a.data)
	return result
}

// Range 遍历数组并对每个元素执行函数
// 如果函数返回false，则停止遍历
func (a *Array[T]) Range(f func(index int, value T) bool) {
	// 先获取数据副本并释放锁，避免回调函数中操作数组导致死锁
	a.mu.RLock()
	data := make([]T, len(a.data))
	copy(data, a.data)
	a.mu.RUnlock()

	// 遍历副本执行回调
	for i, v := range data {
		if !f(i, v) {
			break
		}
	}
}

// Contains 检查数组是否包含指定元素
func (a *Array[T]) Contains(value T) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, v := range a.data {
		if v == value {
			return true
		}
	}
	return false
}

// IndexOf 查找元素在数组中的索引
func (a *Array[T]) IndexOf(value T) int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for i, v := range a.data {
		if v == value {
			return i
		}
	}
	return -1
}

// Replace 替换所有匹配的元素
func (a *Array[T]) Replace(oldVal, newVal T) int {
	a.mu.Lock()
	defer a.mu.Unlock()

	count := 0
	for i, v := range a.data {
		if v == oldVal {
			a.data[i] = newVal
			count++
		}
	}
	return count
}
