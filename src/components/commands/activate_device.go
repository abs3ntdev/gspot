package commands

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/zmb3/spotify/v2"
)

func (c *Commander) activateDevice() (spotify.ID, error) {
	var device *spotify.PlayerDevice
	configDir, _ := os.UserConfigDir()
	if _, err := os.Stat(filepath.Join(configDir, "gospt/device.json")); err == nil {
		deviceFile, err := os.Open(filepath.Join(configDir, "gospt/device.json"))
		if err != nil {
			return "", err
		}
		defer deviceFile.Close()
		deviceValue, err := io.ReadAll(deviceFile)
		if err != nil {
			return "", err
		}
		err = json.Unmarshal(deviceValue, &device)
		if err != nil {
			return "", err
		}
		err = c.Client.TransferPlayback(c.Context, device.ID, true)
		if err != nil {
			return "", err
		}
	} else {
		c.Log.Error("COMMANDER", "failed to activated device", "YOU MUST RUN gospt setdevice FIRST")
	}
	return device.ID, nil
}
