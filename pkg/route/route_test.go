package route

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 测试基本路由创建和删除
func TestRouteCreationAndDeletion(t *testing.T) {
	router := NewRouter[string](64)

	// 测试创建新路由
	route1, err := router.AddRoute("route1")
	assert.NoError(t, err)
	assert.NotNil(t, route1)

	// 测试创建重复路由
	route2, err := router.AddRoute("route1")
	assert.Error(t, err)
	assert.Equal(t, route1, route2)

	// 测试删除存在的路由
	err = router.RemoveRoute("route1")
	assert.NoError(t, err)

	// 测试删除不存在的路由
	err = router.RemoveRoute("nonexistent")
	assert.Error(t, err)
}

// 测试单播消息
func TestUnicast(t *testing.T) {
	router := NewRouter[string](64)
	go router.Run()
	defer router.Stop()

	sender, _ := router.AddRoute("sender")
	receiver, _ := router.AddRoute("receiver")

	var wg sync.WaitGroup
	wg.Add(1)

	// 设置接收处理器
	receiver.HandlerFunc(func(header RoutePacketHeader, data string) {
		defer wg.Done()
		assert.Equal(t, "test message", data)
		assert.Equal(t, "sender", header.Src)
		assert.Equal(t, "receiver", header.Dest)
		assert.Equal(t, RoutePacketTypeUnicast, header.Type)
	})

	// 发送单播消息
	sender.Send("receiver", "test message")

	// 等待消息处理完成
	wg.Wait()
}

// 测试广播消息
func TestBroadcast(t *testing.T) {
	router := NewRouter[string](64)
	go router.Run()
	defer router.Stop()

	sender, _ := router.AddRoute("sender")
	recv1, _ := router.AddRoute("recv1")
	recv2, _ := router.AddRoute("recv2")

	var wg sync.WaitGroup
	wg.Add(2)

	// 设置接收处理器
	recv1.HandlerFunc(func(header RoutePacketHeader, data string) {
		defer wg.Done()
		assert.Equal(t, "broadcast", data)
		assert.Equal(t, "sender", header.Src)
		assert.Equal(t, RoutePacketTypeBroadcast, header.Type)
	})

	recv2.HandlerFunc(func(header RoutePacketHeader, data string) {
		defer wg.Done()
		assert.Equal(t, "broadcast", data)
	})

	// 发送广播消息
	sender.SendBroadcast("broadcast")

	// 等待消息处理完成
	wg.Wait()
}

// 测试组播消息
func TestMulticast(t *testing.T) {
	router := NewRouter[string](64)
	go router.Run()
	defer router.Stop()

	sender, _ := router.AddRoute("sender")
	groupMember1, _ := router.AddRoute("member1")
	groupMember2, _ := router.AddRoute("member2")
	nonMember, _ := router.AddRoute("nonmember")

	// 加入组
	groupMember1.JoinGroup("testgroup")
	groupMember2.JoinGroup("testgroup")

	var wg sync.WaitGroup
	wg.Add(2)

	// 设置接收处理器
	groupMember1.HandlerFunc(func(header RoutePacketHeader, data string) {
		defer wg.Done()
		assert.Equal(t, "multicast", data)
		assert.Equal(t, "testgroup", header.Dest)
		assert.Equal(t, RoutePacketTypeMulticast, header.Type)
	})

	groupMember2.HandlerFunc(func(header RoutePacketHeader, data string) {
		defer wg.Done()
		assert.Equal(t, "multicast", data)
	})

	// 非组成员不应收到消息
	nonMember.HandlerFunc(func(header RoutePacketHeader, data string) {
		t.Error("非组成员不应收到组播消息")
	})

	// 发送组播消息
	sender.SendGroup("testgroup", "multicast")

	// 等待消息处理完成
	wg.Wait()
}

// 测试TTL处理
func TestTTLHandling(t *testing.T) {
	router := NewRouter[string](2) // 设置TTL为2
	go router.Run()
	defer router.Stop()

	routeA, _ := router.AddRoute("A")
	routeB, _ := router.AddRoute("B")
	routeC, _ := router.AddRoute("C")

	// 路由B转发消息到C
	routeB.HandlerFunc(func(header RoutePacketHeader, data string) {
		routeB.SendForward("C", &header, data)
	})

	// 路由C接收消息
	var cReceived bool
	routeC.HandlerFunc(func(header RoutePacketHeader, data string) {
		cReceived = true
	})

	// 发送消息，TTL=2
	routeA.Send("B", "test ttl")

	// 等待消息处理
	time.Sleep(100 * time.Millisecond)

	// 验证C收到了消息（TTL=2: A->B消耗1，B->C消耗1，刚好到达）
	assert.True(t, cReceived)

	// 重置状态并测试TTL耗尽情况
	cReceived = false
	router = NewRouter[string](1) // 设置TTL为1
	go router.Run()
	defer router.Stop()

	routeA, _ = router.AddRoute("A")
	routeB, _ = router.AddRoute("B")
	routeC, _ = router.AddRoute("C")

	routeB.HandlerFunc(func(header RoutePacketHeader, data string) {
		routeB.SendForward("C", &header, data)
	})

	routeC.HandlerFunc(func(header RoutePacketHeader, data string) {
		cReceived = true
	})

	// 发送消息，TTL=1
	routeA.Send("B", "test ttl expired")

	// 等待消息处理
	time.Sleep(100 * time.Millisecond)

	// 验证C没有收到消息（TTL=1: A->B消耗1，B->C时TTL已耗尽）
	assert.False(t, cReceived)
}

// 测试过滤器功能
func TestFilters(t *testing.T) {
	router := NewRouter[string](64)
	go router.Run()
	defer router.Stop()

	sender, _ := router.AddRoute("sender")
	receiver, _ := router.AddRoute("receiver")

	// 创建过滤器 - 阻止特定消息
	blockFilter := func(header RoutePacketHeader, data string) bool {
		return data == "blocked"
	}

	// 添加过滤器
	err := receiver.AddFilter("receiver", (*Filter[string])(&blockFilter))
	assert.NoError(t, err)

	var receivedCount int
	receiver.HandlerFunc(func(header RoutePacketHeader, data string) {
		receivedCount++
	})

	// 发送应被过滤的消息
	sender.Send("receiver", "blocked")

	// 发送应被接收的消息
	sender.Send("receiver", "allowed")

	// 等待消息处理
	time.Sleep(100 * time.Millisecond)

	// 验证只收到了1条消息
	assert.Equal(t, 1, receivedCount)

	// 移除过滤器
	err = receiver.RemoveFilter("receiver", (*Filter[string])(&blockFilter))
	assert.NoError(t, err)

	// 再次发送之前被过滤的消息
	sender.Send("receiver", "blocked")

	// 等待消息处理
	time.Sleep(100 * time.Millisecond)

	// 验证现在收到了2条消息
	assert.Equal(t, 2, receivedCount)
}

// 测试组管理
func TestGroupManagement(t *testing.T) {
	router := NewRouter[string](64)

	route1, _ := router.AddRoute("route1")

	// 加入组
	route1.JoinGroup("group1")
	route1.JoinGroup("group2")

	// 验证组成员关系
	group1, _ := router.groups.Load("group1")
	assert.True(t, group1.Contains("route1"))

	group2, _ := router.groups.Load("group2")
	assert.True(t, group2.Contains("route1"))

	assert.True(t, route1.groups.Contains("group1"))
	assert.True(t, route1.groups.Contains("group2"))

	// 离开组
	route1.LeaveGroup("group1")

	// 验证离开组后的状态
	group1, exists := router.groups.Load("group1")
	assert.False(t, exists)
	assert.False(t, route1.groups.Contains("group1"))

	// 删除路由，验证从所有组中移除
	err := router.RemoveRoute("route1")
	assert.NoError(t, err)

	group2, exists = router.groups.Load("group2")
	assert.False(t, exists)
}

// 测试并发操作
func TestConcurrentOperations(t *testing.T) {
	router := NewRouter[string](64)
	go router.Run()
	defer router.Stop()
	var count atomic.Int32
	var wg sync.WaitGroup

	// 并发创建路由
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			name := fmt.Sprintf("route%d", id)
			route, err := router.AddRoute(name)
			assert.NoError(t, err)
			assert.NotNil(t, route)
			route.HandlerFunc(func(header RoutePacketHeader, data string) {
				count.Add(1)
			})
		}(i)
	}
	wg.Wait()

	// 并发加入组
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			name := fmt.Sprintf("route%d", id)
			if route, ok := router.routes.Load(name); ok {
				route.JoinGroup("concurrent-group")
			}
		}(i)
	}
	wg.Wait()

	// 并发发送消息
	for i := 0; i < 80; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			senderName := fmt.Sprintf("route%d", id%10)
			if sender, ok := router.routes.Load(senderName); ok {
				sender.SendBroadcast(fmt.Sprintf("message %d", id))
			}
		}(i)
	}
	wg.Wait()
	// 等待消息处理完毕
	group, ok := router.groups.Load("concurrent-group")
	assert.True(t, ok)
	assert.Equal(t, 10, group.Cardinality())
	time.Sleep(1000 * time.Millisecond)
	// 发送 80 次广播 ，广播不会传播到自身 (10 个路由发送 80 次，每次有 9 个其他路由接收到广播)
	assert.Equal(t, 720, int(count.Load()))
}
