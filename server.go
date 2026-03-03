package frodding

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/cubexteam/Frodding/logger"
	"github.com/cubexteam/Frodding/network"
	"github.com/cubexteam/Frodding/plugin"
	"github.com/cubexteam/Frodding/resources"
)

const (
	Name    = "Frodding"
	Version = "1.0.0"
	Author  = "SantianDev, bota"
)

type Server struct {
	Config  *resources.Config
	Log     *logger.Logger
	Network *network.Server
	Running bool
}

func NewServer(configPath string) (*Server, error) {
	cfg, err := resources.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	log := logger.New(cfg.Console.Debug)
	netSrv := network.NewServer(cfg, log)
	return &Server{
		Config:  cfg,
		Log:     log,
		Network: netSrv,
	}, nil
}

func (s *Server) Start() error {
	s.printBanner()

	runPlugins, err := plugin.Initialize(s.Log, s.Config.Plugins.Folder)
	if err != nil {
		return fmt.Errorf("plugin init failed: %w", err)
	}

	s.Running = true

	s.Network.SetShutdownHook(func() {
		s.Running = false
	})

	if err := s.Network.Start(); err != nil {
		return err
	}

	ctx, stopPlugins := context.WithCancel(context.Background())
	var wg *sync.WaitGroup
	wg = runPlugins(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for s.Running {
		select {
		case <-sigChan:
			fmt.Println()
			s.Log.Infof("Shutting down...")
			s.Stop()
		case <-ticker.C:
		}
	}

	stopPlugins()
	wg.Wait()
	return nil
}

func (s *Server) Stop() {
	s.Running = false
	s.Network.Shutdown()
	s.Log.Infof("Server stopped.")
}

func (s *Server) printBanner() {
	fmt.Printf("\n")
	fmt.Printf("  ███████╗██████╗  ██████╗ ██████╗ ██████╗ ██╗███╗  ██╗ ██████╗ \n")
	fmt.Printf("  ██╔════╝██╔══██╗██╔═══██╗██╔══██╗██╔══██╗██║████╗ ██║██╔════╝ \n")
	fmt.Printf("  █████╗  ██████╔╝██║   ██║██║  ██║██║  ██║██║██╔██╗██║██║  ███╗\n")
	fmt.Printf("  ██╔══╝  ██╔══██╗██║   ██║██║  ██║██║  ██║██║██║╚████║██║   ██║\n")
	fmt.Printf("  ██║     ██║  ██║╚██████╔╝██████╔╝██████╔╝██║██║ ╚███║╚██████╔╝\n")
	fmt.Printf("  ╚═╝     ╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚═════╝ ╚═╝╚═╝  ╚══╝ ╚═════╝ \n")
	fmt.Printf("\n")
	fmt.Printf("  %s v%s by %s\n", Name, Version, Author)
	fmt.Printf("  MCBE 1.1.x (Protocol 113)\n")
	fmt.Printf("  https://github.com/cubexteam/Frodding\n")
	fmt.Printf("\n")
}
