package plugin

import (
	"context"
	"fmt"
	"os"
)

type Impl interface {
	Name() string
	Setup(p *Plugin) error
	Run(ctx context.Context, p *Plugin)
}

type Logger interface {
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
	Debugf(format string, args ...any)
}

type Plugin struct {
	impl      Impl
	logger    Logger
	directory string
	dirOnce   struct {
		done bool
	}
}

func (p *Plugin) Impl() Impl     { return p.impl }
func (p *Plugin) Logger() Logger { return p.logger }

func (p *Plugin) DataFolder() string {
	if !p.dirOnce.done {
		if err := os.MkdirAll(p.directory, 0o755); err != nil {
			panic(fmt.Sprintf("plugin %s: failed to create data folder: %v", p.impl.Name(), err))
		}
		p.dirOnce.done = true
	}
	return p.directory
}
