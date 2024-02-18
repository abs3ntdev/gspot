package commands

import (
	"os"
	"path/filepath"
)

func (c *Commander) DownloadCover(path string) error {
	if path == "" {
		path = "cover.png"
	}
	destinationPath := filepath.Clean(path)
	state, err := c.Client.PlayerState(c.Context)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(destinationPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	err = state.Item.Album.Images[0].Download(f)
	if err != nil {
		return err
	}
	return nil
}
