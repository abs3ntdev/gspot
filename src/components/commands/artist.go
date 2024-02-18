package commands

import "github.com/zmb3/spotify/v2"

func (c *Commander) ArtistAlbums(artist spotify.ID, page int) (*spotify.SimpleAlbumPage, error) {
	albums, err := c.Client.
		GetArtistAlbums(c.Context, artist, []spotify.AlbumType{1, 2, 3, 4}, spotify.Market(spotify.CountryUSA), spotify.Limit(50), spotify.Offset((page-1)*50))
	if err != nil {
		return nil, err
	}
	return albums, nil
}
