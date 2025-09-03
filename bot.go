package bot

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"sync/atomic"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/greek-milk-bot/core/pkg/core"
	"github.com/greek-milk-bot/core/pkg/utils"
)

type Bot struct {
	plugins map[string]*core.PluginInstance

	channel chan []byte // 数据包路由

	self *PluginSelf // 自我交互用组件

	started *atomic.Bool // 启动控制

	groups *utils.Map[string, mapset.Set[string]]
}

func NewBot(calls ...ConfigFunc) (*Bot, error) {
	config := NewConfig()
	for _, call := range calls {
		if err := call(config); err != nil {
			return nil, err
		}
	}
	init := &Bot{
		started: new(atomic.Bool),
		plugins: make(map[string]*core.PluginInstance),
		self:    NewPluginSelf(),
		groups:  utils.NewMap[string, mapset.Set[string]](),
	}
	init.channel = make(chan []byte, config.Cache)

	init.plugins["self"] = core.NewPluginInstance(init.self)
	for _, plugin := range config.Plugins {
		id, err := uuid.NewUUID()
		if err != nil {
			return nil, err
		}
		init.plugins[id.String()] = core.NewPluginInstance(plugin)
	}
	return init, nil
}

func (g *Bot) Run(ctx context.Context) error {
	if g.started.Swap(true) {
		return errors.New("already started")
	}
	bootCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	for pluginID, plugin := range g.plugins {
		plugin.Bus = core.NewPluginBus(pluginID, bootCtx)
		err := plugin.Boot(plugin.Bus)
		if err != nil {
			return err
		}
	}
	for {
		select {
		case <-bootCtx.Done():
			return nil
		case msg := <-g.channel:
			go func() {
				if err := g.eval(msg); err != nil {
					log.Printf("error: %s\n\nMessage:%s", err, string(msg))
				}
			}()
		}
	}
}

func (g *Bot) eval(msg []byte) error {
	var packet core.Packet
	if err := json.Unmarshal(msg, &packet); err != nil {
		return err
	}
	switch packet.Type {
	case core.PacketTypeMeta:
		if err := g.evalMeta(packet.Src, packet.Data.(*core.PacketMeta)); err != nil {
			return err
		}
	case core.PacketTypeEvent:
		group, isGroupEvent := g.broadCastGroup(packet)
		if isGroupEvent {
			// 将路由事件转移到其他事件
			for id := range group {
				packet.Dest = id
				data, err := json.Marshal(packet)
				if err != nil {
					return err
				}
				g.channel <- data
			}
		} else if plugin, ok := g.plugins[packet.Dest]; ok {
			// 处理单条事件
			return g.evalEvent(plugin, packet.Src, packet.Dest, packet.Data.(*core.PacketEvent))
		}
		return nil
	case core.PacketTypeCall:

	}
	return nil
}

func (g *Bot) evalEvent(plugin *core.PluginInstance, src string, dest string, event *core.PacketEvent) error {
	return nil
}

func (g *Bot) broadCastGroup(packet core.Packet) (map[string]*core.PluginInstance, bool) {
	results := make(map[string]*core.PluginInstance)
	var isGroup bool
	if packet.Dest == "broadcast" {
		for id, instance := range g.plugins {
			results[id] = instance
		}
		isGroup = true
	} else if strings.HasPrefix(packet.Dest, "group:") {
		group := strings.TrimPrefix(packet.Dest, "group:")
		if value, ok := g.groups.Load(group); ok {
			for id := range value.(mapset.Set[string]).Iter() {
				results[id] = g.plugins[id]
			}
		}
		isGroup = true
	}
	return results, isGroup
}

func (g *Bot) evalMeta(src string, meta *core.PacketMeta) error {
	switch meta.Action {
	case "group::bind":
		if meta.Data != "" {
			actual, _ := g.groups.LoadOrStore(meta.Data, mapset.NewSet[string]())
			actual.(mapset.Set[string]).Add(src)
		}
	}
	return nil
}
