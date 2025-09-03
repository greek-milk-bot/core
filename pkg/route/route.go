package route

import (
	"errors"
	"fmt"
	"sync"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/greek-milk-bot/core/pkg/utils"
)

type Router[T any] struct {
	defaultTtl uint8 // 默认TTL值
	routes     *utils.Map[string, *Route[T]]
	messages   chan RoutePacket[T]
	groups     *utils.Map[string, mapset.Set[string]]                           // 组-成员映射
	filter     *utils.Map[string, *utils.Map[string, *utils.Array[*Filter[T]]]] // 过滤器

	once sync.Once
}

func NewRouter[T any](ttl uint8) *Router[T] {
	if ttl == 0 {
		ttl = 64
	}
	return &Router[T]{
		defaultTtl: ttl,
		messages:   make(chan RoutePacket[T], ttl),
		filter:     utils.NewMap[string, *utils.Map[string, *utils.Array[*Filter[T]]]](),
		routes:     utils.NewMap[string, *Route[T]](),
		groups:     utils.NewMap[string, mapset.Set[string]](),

		once: sync.Once{},
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

	groups mapset.Set[string] // 该路由加入的组
}

func (r *Router[T]) AddRoute(name string) (*Route[T], error) {
	store, loaded := r.routes.LoadOrStore(name, &Route[T]{
		router:  r,
		name:    name,
		handler: nil,
		groups:  mapset.NewSet[string](),
	})
	if loaded {
		return store, fmt.Errorf("route %s already exists", name)
	}
	return store, nil
}

func (r *Router[T]) RemoveRoute(name string) error {
	item, ok := r.routes.LoadAndDelete(name)
	if !ok {
		return fmt.Errorf("route %s not found", name)
	}

	// 第一步：遍历所有 pattern，删除该路由的过滤器条目，收集需检查的 pattern
	var emptyPatternCandidates []string
	r.filter.Range(func(pattern string, innerMap *utils.Map[string, *utils.Array[*Filter[T]]]) bool {
		// 删除内层 Map 中该路由的过滤器条目
		deleted := innerMap.RemoveIf(name, func(arr *utils.Array[*Filter[T]]) bool {
			return true // 无论数组是否为空，都删除该路由的条目
		})
		// 若删除了条目，标记该 pattern 需后续检查是否为空
		if deleted {
			emptyPatternCandidates = append(emptyPatternCandidates, pattern)
		}
		return true // 继续遍历所有 pattern
	})

	// 第二步：检查收集的 pattern，若 innerMap 为空则删除
	for _, pattern := range emptyPatternCandidates {
		// 重新加载 innerMap（避免 Range 中引用失效）
		innerMap, ok := r.filter.Load(pattern)
		if !ok {
			continue
		}
		// 若 innerMap 为空，删除该 pattern
		if innerMap.Len() == 0 {
			r.filter.LoadAndDelete(pattern)
		}
	}

	// 从所有组中移除
	for _, group := range item.groups.ToSlice() {
		item.LeaveGroup(group)
	}
	return nil
}

type (
	Handler[T any] func(header RoutePacketHeader, data T)
	Filter[T any]  func(header RoutePacketHeader, data T) bool
)

// 发送单播包
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
	if stack.Ttl <= 1 {
		return // TTL即将耗尽，不再转发
	}

	newStack := make([]string, len(stack.Stack))
	copy(newStack, stack.Stack)
	newStack = append(newStack, r.name)

	r.router.messages <- RoutePacket[T]{
		Header: RoutePacketHeader{
			Type:  stack.Type,
			Src:   stack.Src,
			Dest:  dest,
			Stack: newStack,
			Ttl:   stack.Ttl - 1,
		},
		Data: message,
	}
}

// 发送广播包
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
	// 将路由自身添加到组的成员集合
	groupSet, _ := r.router.groups.LoadOrStore(group, mapset.NewSet[string]())
	groupSet.Add(r.name)
	// 记录路由加入的组
	r.groups.Add(group)
}

// 离开组
func (r *Route[T]) LeaveGroup(group string) {
	if groupSet, found := r.router.groups.Load(group); found {
		groupSet.Remove(r.name)
		// 如果组为空则删除组
		r.router.groups.RemoveIf(group, func(s mapset.Set[string]) bool {
			return s.IsEmpty()
		})
	}
	// 从路由的组列表中移除
	r.groups.Remove(group)
}

// AddFilter 添加过滤器
func (r *Route[T]) AddFilter(pattern string, handler *Filter[T]) error {
	grp, _ := r.router.filter.LoadOrStore(pattern, utils.NewMap[string, *utils.Array[*Filter[T]]]())
	filter, _ := grp.LoadOrStore(r.name, utils.NewArray[*Filter[T]]())
	if !filter.AddIfNotExists(handler) {
		return errors.New("filter already exists")
	}
	return nil
}

func (r *Route[T]) RemoveFilter(pattern string, handler *Filter[T]) error {
	if grp, find := r.router.filter.Load(pattern); find {
		if filter, find := grp.Load(r.name); find {
			filter.DeleteByValue(handler)
			r.router.filter.RemoveIf(pattern, func(u *utils.Map[string, *utils.Array[*Filter[T]]]) bool {
				u.RemoveIf(r.name, func(u *utils.Array[*Filter[T]]) bool {
					return u.Len() == 0
				})
				return u.Len() == 0
			})
		}
	}
	return nil
}

func (r *Route[T]) HandlerFunc(handler Handler[T]) {
	r.handler = handler
}

func (r *Router[T]) Run() {
	for packet := range r.messages {
		// TTL检查
		if packet.Header.Ttl <= 0 {
			continue
		}
		// 过滤器处理
		filter, hasFilter := r.filter.Load(packet.Header.Dest)
		if hasFilter && filter != nil {
			finish := false
			filters := make([]*Filter[T], 0)
			filter.Range(func(key string, value *utils.Array[*Filter[T]]) bool {
				filters = append(filters, value.Slice()...)
				return true
			})
			for _, item := range filters {
				f := *item
				if f(packet.Header, packet.Data) {
					finish = true
					break
				}
			}
			if finish {
				// 包已经被拦截，跳过
				continue
			}
		}
		// 根据包类型分发
		switch packet.Header.Type {
		case RoutePacketTypeUnicast:
			r.handleUnicast(packet)
		case RoutePacketTypeBroadcast:
			r.handleBroadcast(packet)
		case RoutePacketTypeMulticast:
			r.handleMulticast(packet)
		}
	}
}

func (r *Router[T]) handleUnicast(packet RoutePacket[T]) {
	if destRoute, ok := r.routes.Load(packet.Header.Dest); ok && destRoute.handler != nil {
		destRoute.handler(packet.Header, packet.Data)
	}
}

// 处理广播消息
func (r *Router[T]) handleBroadcast(packet RoutePacket[T]) {
	r.routes.Range(func(name string, route *Route[T]) bool {
		// 不向发送者自身广播
		if name != packet.Header.Src && route.handler != nil {
			route.handler(packet.Header, packet.Data)
		}
		return true
	})
}

// 处理组播消息
func (r *Router[T]) handleMulticast(packet RoutePacket[T]) {
	if groupSet, ok := r.groups.Load(packet.Header.Dest); ok {
		for _, memberName := range groupSet.ToSlice() {
			// 不向发送者自身组播
			if memberName != packet.Header.Src {
				if memberRoute, ok := r.routes.Load(memberName); ok && memberRoute.handler != nil {
					memberRoute.handler(packet.Header, packet.Data)
				}
			}
		}
	}
}

func (r *Router[T]) Stop() {
	r.once.Do(func() {
		close(r.messages)
	})
}
