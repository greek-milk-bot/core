package core

import (
	"context"
)

type PluginBus struct {
	context.Context
	ID string
}

func NewPluginBus(id string, ctx context.Context) *PluginBus {
	return &PluginBus{
		Context: ctx,
		ID:      id,
	}
}
