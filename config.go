package bot

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/greek-milk-bot/core/pkg/core"
)

type ConfigFunc func(*Config) error

type Config struct {
	Plugins []core.Plugin
	Cache   int
}

func NewConfig() *Config {
	return &Config{
		Plugins: make([]core.Plugin, 0),
		Cache:   100,
	}
}

func WithPlugins(plugins ...core.Plugin) ConfigFunc {
	return func(config *Config) error {
		for _, plugin := range plugins {
			find := false
			for _, plugin := range config.Plugins {
				if plugin != plugin {
					find = true
					break
				}
			}
			if !find {
				config.Plugins = append(config.Plugins, plugin)
			}
		}
		return nil
	}
}

func WithPluginURL(ctx context.Context, urlStr string) ConfigFunc {
	return func(config *Config) error {
		sType, urlStr, found := strings.Cut(urlStr, "+")
		if !found {
			return fmt.Errorf("invalid url: %s", urlStr)
		}
		adapter, ok := core.GetPlugin(sType)
		if !ok {
			return fmt.Errorf("adapter not found: %s", sType)
		}
		u, err := url.Parse(urlStr)
		if err != nil {
			return err
		}
		b, err := adapter(ctx, *u)
		if err != nil {
			return err
		}
		return WithPlugins(b)(config)
	}
}
