package commands

import "github.com/zmb3/spotify/v2"

func (c *Commander) Playlists(page int) (*spotify.SimplePlaylistPage, error) {
	return c.Client().CurrentUsersPlaylists(c.Context, spotify.Limit(50), spotify.Offset((page-1)*50))
}

func (c *Commander) PlaylistTracks(playlist spotify.ID, page int) (*spotify.PlaylistItemPage, error) {
	return c.Client().GetPlaylistItems(c.Context, playlist, spotify.Limit(50), spotify.Offset((page-1)*50))
}

func (c *Commander) DeleteTracksFromPlaylist(tracks []spotify.ID, playlist spotify.ID) error {
	_, err := c.Client().RemoveTracksFromPlaylist(c.Context, playlist, tracks...)
	if err != nil {
		return err
	}
	return nil
}

func (c *Commander) TrackList(page int) (*spotify.SavedTrackPage, error) {
	return c.Client().CurrentUsersTracks(c.Context, spotify.Limit(50), spotify.Offset((page-1)*50))
}
