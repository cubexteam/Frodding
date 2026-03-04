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
	initialized atomic.Bool
	plugins     []*Plugin
)

func Add(impl Impl) {
	if initialized.Load() {
		panic("cannot add plugin after server has started")
	}
	plugins = append(plugins, &Plugin{impl: impl})
}

func Initialize(log Logger, folder string) (func(context.Context) *sync.WaitGroup, error) {
	if !initialized.CompareAndSwap(false, true) {
		panic("plugins already initialized")
	}

	log.Infof("Loading %d plugin(s)...", len(plugins))

	seen := make(map[string]struct{}, len(plugins))
	for _, p := range plugins {
		key := strings.ToLower(p.impl.Name())
		if _, dup := seen[key]; dup {
			return nil, fmt.Errorf("duplicate plugin name: %s", p.impl.Name())
		}
		seen[key] = struct{}{}
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	for _, p := range plugins {
		p.logger = log
		p.directory = filepath.Join(wd, folder, p.impl.Name())
		if err := p.impl.Setup(p); err != nil {
			return nil, fmt.Errorf("plugin %q setup failed: %w", p.impl.Name(), err)
		}
		log.Infof("Plugin %q loaded.", p.impl.Name())
	}

	return func(ctx context.Context) *sync.WaitGroup {
		var wg sync.WaitGroup
		wg.Add(len(plugins))
		for _, p := range plugins {
			go func(p *Plugin) {
				defer wg.Done()
				p.impl.Run(ctx, p)
			}(p)
		}
		return &wg
	}, nil
}
