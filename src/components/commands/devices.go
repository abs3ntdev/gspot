package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zmb3/spotify/v2"
)

func (c *Commander) ListDevices() error {
	devices, err := c.Client().PlayerDevices(c.Context)
	if err != nil {
		return err
	}
	return PrintDevices(devices)
}

func PrintDevices(devices []spotify.PlayerDevice) error {
	out, err := json.MarshalIndent(devices, "", " ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func (c *Commander) SetDevice(device spotify.ID) error {
	err := c.Client().TransferPlayback(c.Context, device, true)
	if err != nil {
		return err
	}
	devices, err := c.Client().PlayerDevices(c.Context)
	if err != nil {
		return err
	}
	for _, d := range devices {
		if d.ID == device {
			out, err := json.MarshalIndent(d, "", " ")
			if err != nil {
				return err
			}
			configDir, _ := os.UserConfigDir()
			err = os.WriteFile(filepath.Join(configDir, "gspot/device.json"), out, 0o600)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("device not found")
}
