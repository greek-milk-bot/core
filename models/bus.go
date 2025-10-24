package models

import (
	"context"
)

type PluginBus struct {
	context.Context
	ID string
}

func (bus PluginBus) SendPacket(packet Packet) error {
	panic("implement me")
}
