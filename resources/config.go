package resources

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type ConsoleConfig struct {
	Debug bool `yaml:"debug"`
}

type PluginsConfig struct {
	Folder string `yaml:"folder"`
}

type Config struct {
	ServerName   string        `yaml:"server-name"`
	ServerIP     string        `yaml:"server-ip"`
	ServerPort   int           `yaml:"server-port"`
	MaxPlayers   int           `yaml:"max-players"`
	MOTD         string        `yaml:"motd"`
	MotdProtocol int           `yaml:"motd-protocol"`
	Console      ConsoleConfig `yaml:"console"`
	Plugins      PluginsConfig `yaml:"plugins"`
}

func DefaultConfig() *Config {
	return &Config{
		ServerName:   "Frodding Server",
		ServerIP:     "0.0.0.0",
		ServerPort:   19132,
		MaxPlayers:   20,
		MOTD:         "A Frodding Server",
		Console:      ConsoleConfig{Debug: false},
		Plugins:      PluginsConfig{Folder: "plugins"},
	}
}

func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if saveErr := SaveConfig(path, cfg); saveErr != nil {
				fmt.Printf("[Frodding] Warning: could not save default config: %v\n", saveErr)
			} else {
				fmt.Println("[Frodding] Created default server.yml")
			}
			return cfg, nil
		}
		return nil, err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func SaveConfig(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
