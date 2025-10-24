package bot

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/greek-milk-bot/core/models"
	"github.com/greek-milk-bot/core/utils"
)

type GreekMilkBot struct {
	plugins map[string]*models.PluginInstance
	route   *Router[models.Packet]
	once    *atomic.Bool
}

func NewGreekMilkBot(plugins ...models.Plugin) (*GreekMilkBot, error) {
	if len(plugins) == 0 {
		return nil, errors.New("no plugins")
	}
	r := &GreekMilkBot{
		plugins: make(map[string]*models.PluginInstance),
		route:   NewRouter[models.Packet](8),
		once:    new(atomic.Bool),
	}
	for i, plugin := range plugins {
		id := fmt.Sprintf("%d", i)
		if plugin == nil {
			return nil, errors.New("nil plugin")
		}
		inst := &models.PluginInstance{
			Plugin:    plugin,
			Meta:      utils.NewMap[string, string](),
			Tools:     mapset.NewSet[string](),
			Resources: utils.NewMap[string, models.ResourceProvider](),
		}
		r.plugins[id] = inst
	}
	return r, nil
}
func (r *GreekMilkBot) Run(ctx context.Context) error {
	if r.once.Swap(true) {
		return errors.New("plugin already running")
	}
	for id, plugin := range r.plugins {
		if err := func() error {
			bootCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			plugin.Bus = models.PluginBus{
				Context: bootCtx,
				ID:      id,
			}
			if err := plugin.Boot(plugin.Bus); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return err
		}
	}
	r.route.RunContext(ctx)
	return nil
}
