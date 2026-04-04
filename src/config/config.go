package config

import (
	"os"

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

	Librespot LibrespotConfig `yaml:"librespot"`
}

type LibrespotConfig struct {
	Enabled      bool   `yaml:"enabled"`
	DeviceName   string `yaml:"device_name"`
	AudioBackend string `yaml:"audio_backend" default:"pulseaudio"`
	AudioDevice  string `yaml:"audio_device"  default:"default"`
	Bitrate      int    `yaml:"bitrate"       default:"160"`
	VolumeSteps  uint32 `yaml:"volume_steps"  default:"100"`
	InitialVol   uint32 `yaml:"initial_volume" default:"100"`
	Normalise    bool   `yaml:"normalisation"`
	Zeroconf     bool   `yaml:"zeroconf"      default:"true"`
	Autoplay     bool   `yaml:"autoplay"      default:"true"`
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
	if c.Librespot.DeviceName == "" {
		hostname, _ := os.Hostname()
		if hostname != "" {
			c.Librespot.DeviceName = "gspot-" + hostname
		} else {
			c.Librespot.DeviceName = "gspot"
		}
	}
	if c.Librespot.AudioBackend == "" {
		c.Librespot.AudioBackend = "pulseaudio"
	}
	if c.Librespot.AudioDevice == "" {
		c.Librespot.AudioDevice = "default"
	}
	if c.Librespot.Bitrate == 0 {
		c.Librespot.Bitrate = 160
	}
	if c.Librespot.VolumeSteps == 0 {
		c.Librespot.VolumeSteps = 100
	}
	if c.Librespot.InitialVol == 0 {
		c.Librespot.InitialVol = 100
	}
}

// ConfigDir returns the gspot configuration directory.
func ConfigDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "/tmp"
	}
	return dir + "/gspot"
}
