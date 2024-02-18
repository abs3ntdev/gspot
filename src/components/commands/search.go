package commands

import "github.com/zmb3/spotify/v2"

func (c *Commander) Search(search string, page int) (*spotify.SearchResult, error) {
	result, err := c.Client.
		Search(c.Context, search, spotify.SearchTypeAlbum|spotify.SearchTypeArtist|spotify.SearchTypeTrack|spotify.SearchTypePlaylist, spotify.Limit(50), spotify.Offset((page-1)*50))
	if err != nil {
		return nil, err
	}
	return result, nil
}
