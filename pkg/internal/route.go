package internal

import (
	"errors"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/greek-milk-bot/core/pkg/utils"
)

type Router[T any] struct {
	defaultTtl uint8 // 默认丢弃大小
	routes     *utils.Map[string, *Route[T]]
	messages   chan RoutePacket[T]
	filter     *utils.Map[string, Handler[T]] // todo: 处理数据销毁
	groups     *utils.Map[string, mapset.Set[string]]
}

func NewRouter[T any](ttl uint8) *Router[T] {
	if ttl == 0 {
		ttl = 64
	}
	return &Router[T]{
		defaultTtl: ttl,
		messages:   make(chan RoutePacket[T], ttl),
		filter:     utils.NewMap[string, Handler[T]](),
		routes:     utils.NewMap[string, *Route[T]](),
		groups:     utils.NewMap[string, mapset.Set[string]](),
	}
}

type RoutePacketType uint8

const (
	RoutePacketTypeUnicast RoutePacketType = iota
	RoutePacketTypeBroadcast
	RoutePacketTypeMulticast
)

type RoutePacket[T any] struct {
	Header RoutePacketHeader
	Data   T
}

type RoutePacketHeader struct {
	Type  RoutePacketType
	Src   string
	Dest  string
	Stack []string
	Ttl   uint8
}

type Route[T any] struct {
	name    string
	router  *Router[T]
	handler Handler[T]

	groups mapset.Set[string]
}

func (r *Router[T]) AddRoute(name string) (*Route[T], error) {
	store, b := r.routes.LoadOrStore(name, &Route[T]{
		router:  r,
		name:    name,
		handler: nil,
		groups:  mapset.NewSet[string](),
	})
	if b {
		return store, fmt.Errorf("route %s already exists", name)
	}
	return store, nil
}

func (r *Router[T]) RemoveRoute(name string) error {
	item, ok := r.routes.LoadAndDelete(name)
	if !ok {
		return fmt.Errorf("route %s not found", name)
	}
	groups := item.groups.ToSlice()
	for _, group := range groups {
		item.LeaveGroup(group)
	}
	return nil
}

type Handler[T any] func(header RoutePacketHeader, data T)

// 发送包
func (r *Route[T]) Send(dest string, message T) {
	r.router.messages <- RoutePacket[T]{
		Header: RoutePacketHeader{
			Type:  RoutePacketTypeUnicast,
			Src:   r.name,
			Dest:  dest,
			Stack: []string{r.name},
			Ttl:   r.router.defaultTtl,
		},
		Data: message,
	}
}

// 发送转发包
func (r *Route[T]) SendForward(dest string, stack *RoutePacketHeader, message T) {
	route := RoutePacketHeader{
		Type:  stack.Type,
		Src:   stack.Src,
		Dest:  dest,
		Stack: stack.Stack,
		Ttl:   stack.Ttl - 1,
	}
	route.Stack = append(route.Stack, r.name)
	r.router.messages <- RoutePacket[T]{
		Header: route,
		Data:   message,
	}
}

// 发送广播
func (r *Route[T]) SendBroadcast(message T) {
	r.router.messages <- RoutePacket[T]{
		Header: RoutePacketHeader{
			Type:  RoutePacketTypeBroadcast,
			Src:   r.name,
			Dest:  "",
			Stack: []string{r.name},
			Ttl:   r.router.defaultTtl,
		},
		Data: message,
	}
}

// 发送组播包
func (r *Route[T]) SendGroup(group string, message T) {
	r.router.messages <- RoutePacket[T]{
		Header: RoutePacketHeader{
			Type:  RoutePacketTypeMulticast,
			Src:   r.name,
			Dest:  group,
			Stack: []string{r.name},
			Ttl:   r.router.defaultTtl,
		},
		Data: message,
	}
}

// 加入组
func (r *Route[T]) JoinGroup(group string) {
	store, _ := r.router.groups.LoadOrStore(group, mapset.NewSet[string]())
	store.Add(group)
	r.groups.Add(group)
}

// 离开组
func (r *Route[T]) LeaveGroup(group string) {
	store, find := r.router.groups.Load(group)
	if find {
		store.Remove(group)
	}
	r.router.groups.RemoveIf(group, func(group mapset.Set[string]) bool {
		return group.IsEmpty()
	})
	r.groups.Remove(group)
}

// 拦截包
func (r *Route[T]) AddFilter(pattern string, handler Handler[T]) error {
	_, b := r.router.filter.LoadOrStore(pattern, handler)
	if b {
		return errors.New("filter already exists")
	}
	return nil
}

// 取消拦截包
func (r *Route[T]) RemoveFilter(pattern string) error {
	_, ok := r.router.filter.LoadAndDelete(pattern)
	if !ok {
		return fmt.Errorf("route %s not found", pattern)
	}
	return nil
}

// 处理接收
func (r *Route[T]) HandlerFunc(handler Handler[T]) {
	r.handler = handler
}
