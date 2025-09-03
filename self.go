package bot

import (
	"github.com/greek-milk-bot/core/pkg/core"
)

type PluginSelf struct{}

func (p *PluginSelf) Boot(ctx *core.PluginBus) error {
	// TODO implement me
	panic("implement me")
}

func NewPluginSelf() *PluginSelf {
	return &PluginSelf{}
}
