package core

import (
	"context"
	"net/url"
)

type PluginHandler func(ctx context.Context, url url.URL) (Plugin, error)

var plugins = make(map[string]PluginHandler)

func GetPlugin(key string) (PluginHandler, bool) {
	handler, ok := plugins[key]
	return handler, ok
}

func RegisterPlugin(name string, plugin PluginHandler) {
	if plugins[name] != nil {
		panic("duplicate adapter name: " + name)
	}
	plugins[name] = plugin
}
