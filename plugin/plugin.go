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

type Plugin struct {
	impl             Impl
	logger           Logger
	directory        string
	directoryCreated bool
}

type Logger interface {
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
	Debugf(format string, args ...any)
}

func (p *Plugin) Impl() Impl      { return p.impl }
func (p *Plugin) Logger() Logger  { return p.logger }

func (p *Plugin) DataFolder() string {
	if !p.directoryCreated {
		if err := os.MkdirAll(p.directory, os.ModePerm); err != nil {
			panic(fmt.Sprintf("failed to create data folder for plugin %s: %v", p.impl.Name(), err))
		}
		p.directoryCreated = true
	}
	return p.directory
}
