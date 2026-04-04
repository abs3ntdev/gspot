package config

import (
	"github.com/adrg/xdg"
)

type Config struct {
	ClientID        string `yaml:"client_id"`
	ClientSecret    string `yaml:"client_secret"`
	ClientSecretCmd string `yaml:"client_secret_cmd"`
	Port            string `yaml:"port"`
	LogLevel        string `yaml:"log_level"         default:"info"`
	LogOutput       string `yaml:"log_output"        default:"stdout"`
	SocketPath      string `yaml:"socket_path"`
	PidFile         string `yaml:"pid_file"`
}

// ResolveDefaults fills in empty fields with XDG-based defaults.
// Called after config loading since these defaults require runtime resolution.
func (c *Config) ResolveDefaults() {
	if c.SocketPath == "" {
		p, err := xdg.RuntimeFile("gspot/gspot.sock")
		if err != nil {
			p = "/tmp/gspot.sock"
		}
		c.SocketPath = p
	}
	if c.PidFile == "" {
		p, err := xdg.RuntimeFile("gspot/gspot.pid")
		if err != nil {
			p = "/tmp/gspot.pid"
		}
		c.PidFile = p
	}
}
