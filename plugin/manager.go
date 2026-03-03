package plugin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	loaded  atomic.Bool
	plugins []*Plugin
)

func Add(impl Impl) {
	if loaded.Load() {
		panic("cannot add plugin after server has started")
	}
	plugins = append(plugins, &Plugin{impl: impl})
}

func Initialize(log Logger, folder string) (func(context.Context) *sync.WaitGroup, error) {
	if !loaded.CompareAndSwap(false, true) {
		panic("plugins already initialized")
	}

	log.Infof("Loading %d plugin(s)...", len(plugins))

	names := map[string]struct{}{}
	for _, p := range plugins {
		key := strings.ToLower(p.impl.Name())
		if _, ok := names[key]; ok {
			return nil, fmt.Errorf("duplicate plugin name: %s", p.impl.Name())
		}
		names[key] = struct{}{}
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	for _, p := range plugins {
		p.logger = log
		p.directory = filepath.Join(wd, folder, p.impl.Name())
		if err := p.impl.Setup(p); err != nil {
			return nil, fmt.Errorf("plugin '%s' setup failed: %w", p.impl.Name(), err)
		}
		log.Infof("Plugin '%s' loaded.", p.impl.Name())
	}

	return func(ctx context.Context) *sync.WaitGroup {
		wg := &sync.WaitGroup{}
		for _, p := range plugins {
			wg.Add(1)
			go func(p *Plugin) {
				defer wg.Done()
				p.impl.Run(ctx, p)
			}(p)
		}
		return wg
	}, nil
}
