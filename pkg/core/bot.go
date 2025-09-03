package core

import (
	"sync"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/greek-milk-bot/core/pkg/utils"
)

type PluginInstance struct {
	Plugin
	Bus       *PluginBus                 // 每个消息通信上下文
	Meta      *utils.Map[string, string] // 元数据
	Tools     mapset.Set[string]         // 可用的 RPC 工具包
	Resources *sync.Map                  // 支持的资源解析器
}

func NewPluginInstance(p Plugin) *PluginInstance {
	return &PluginInstance{
		Plugin:    p,
		Meta:      utils.NewMap[string, string](),
		Tools:     mapset.NewSet[string](),
		Resources: new(sync.Map),
	}
}
