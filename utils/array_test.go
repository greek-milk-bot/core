package utils

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArray_BasicOps(t *testing.T) {
	arr := NewArray[int]()
	assert.Equal(t, 0, arr.Len())
	assert.Equal(t, 0, arr.Cap())

	// Append
	arr.Append(1, 2, 3)
	assert.Equal(t, 3, arr.Len())
	assert.GreaterOrEqual(t, arr.Cap(), 3)

	// Get
	val, err := arr.Get(1)
	assert.NoError(t, err)
	assert.Equal(t, 2, val)

	// Get out of range
	_, err = arr.Get(3)
	assert.Error(t, err)

	// Set
	err = arr.Set(1, 10)
	assert.NoError(t, err)
	val, _ = arr.Get(1)
	assert.Equal(t, 10, val)

	// Set out of range
	err = arr.Set(3, 20)
	assert.Error(t, err)

	// Insert
	err = arr.Insert(2, 20)
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 10, 20, 3}, arr.Slice())

	// Insert out of range
	err = arr.Insert(5, 30)
	assert.Error(t, err)

	// Delete
	delVal, err := arr.Delete(2)
	assert.NoError(t, err)
	assert.Equal(t, 20, delVal)
	assert.Equal(t, []int{1, 10, 3}, arr.Slice())

	// Delete out of range
	_, err = arr.Delete(3)
	assert.Error(t, err)

	// AddIfNotExists
	ok := arr.AddIfNotExists(10)
	assert.False(t, ok) // 10 已存在
	ok = arr.AddIfNotExists(40)
	assert.True(t, ok)
	assert.Equal(t, []int{1, 10, 3, 40}, arr.Slice())

	// Contains
	assert.True(t, arr.Contains(3))
	assert.False(t, arr.Contains(50))

	// IndexOf
	assert.Equal(t, 1, arr.IndexOf(10))
	assert.Equal(t, -1, arr.IndexOf(50))

	// Replace
	count := arr.Replace(10, 100)
	assert.Equal(t, 1, count)
	assert.Equal(t, []int{1, 100, 3, 40}, arr.Slice())

	// Range
	sum := 0
	arr.Range(func(i int, v int) bool {
		sum += v
		return true
	})
	assert.Equal(t, 1+100+3+40, sum)

	// Clear
	arr.Clear()
	assert.Equal(t, 0, arr.Len())
	assert.GreaterOrEqual(t, arr.Cap(), 4) // 保留容量
}

// 并发 Append 和 Get 测试
func TestConcurrentAppendAndGet(t *testing.T) {
	goroutines := 10       // 10个并发goroutine
	opsPerGoroutine := 100 // 每个goroutine执行100次操作

	arr := NewArray[int]()
	var wg sync.WaitGroup

	// 用于记录所有应该被添加的值
	allValues := make([][]int, goroutines)
	for i := range allValues {
		allValues[i] = make([]int, opsPerGoroutine)
	}

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				val := goroutineID*opsPerGoroutine + j
				allValues[goroutineID][j] = val
				arr.Append(val)
			}
		}(i)
	}

	// 等待所有Append操作完成
	wg.Wait()

	// 验证总长度正确
	totalOps := goroutines * opsPerGoroutine
	assert.Equal(t, totalOps, arr.Len())

	// 收集所有添加的值并验证
	values := make(map[int]bool)
	for _, goroutineValues := range allValues {
		for _, val := range goroutineValues {
			values[val] = true
		}
	}

	// 验证每个值都存在于数组中
	arr.Range(func(index int, value int) bool {
		assert.True(t, values[value])
		delete(values, value) // 防止重复验证
		return true
	})

	// 确保没有遗漏任何值
	assert.Empty(t, values)
}
