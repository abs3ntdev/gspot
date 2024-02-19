package commands

import (
	"github.com/zmb3/spotify/v2"
)

func (c *Commander) AlbumTracks(album spotify.ID, page int) (*spotify.SimpleTrackPage, error) {
	tracks, err := c.Client().
		GetAlbumTracks(c.Context, album, spotify.Limit(50), spotify.Offset((page-1)*50), spotify.Market(spotify.CountryUSA))
	if err != nil {
		return nil, err
	}
	return tracks, nil
}

func (c *Commander) UserAlbums(page int) (*spotify.SavedAlbumPage, error) {
	return c.Client().CurrentUsersAlbums(c.Context, spotify.Limit(50), spotify.Offset((page-1)*50))
}
