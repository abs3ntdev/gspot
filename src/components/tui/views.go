package tui

import (
	"fmt"
	"regexp"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/sync/errgroup"

	"git.asdf.cafe/abs3nt/gspot/src/components/commands"
)

const regex = `<.*?>`

func DeviceView(commands *commands.Commander) ([]list.Item, error) {
	items := []list.Item{}
	devices, err := commands.Client().PlayerDevices(commands.Context)
	if err != nil {
		return nil, err
	}
	for _, device := range devices {
		items = append(items, mainItem{
			Name:        device.Name,
			Desc:        fmt.Sprintf("%s - active: %t", device.ID, device.Active),
			SpotifyItem: device,
		})
	}
	return items, nil
}

func QueueView(commands *commands.Commander) ([]list.Item, error) {
	items := []list.Item{}
	tracks, err := commands.Client().GetQueue(commands.Context)
	if err != nil {
		return nil, err
	}
	if tracks.CurrentlyPlaying.Name != "" {
		items = append(items, mainItem{
			Name:     tracks.CurrentlyPlaying.Name,
			Artist:   tracks.CurrentlyPlaying.Artists[0],
			Duration: tracks.CurrentlyPlaying.TimeDuration().Round(time.Second).String(),
			ID:       tracks.CurrentlyPlaying.ID,
			Desc: tracks.CurrentlyPlaying.Artists[0].Name + " - " + tracks.CurrentlyPlaying.TimeDuration().
				Round(time.Second).
				String(),
			SpotifyItem: tracks.CurrentlyPlaying,
		})
	}
	for _, track := range tracks.Items {
		items = append(items, mainItem{
			Name:        track.Name,
			Artist:      track.Artists[0],
			Duration:    track.TimeDuration().Round(time.Second).String(),
			ID:          track.ID,
			Desc:        track.Artists[0].Name + " - " + track.TimeDuration().Round(time.Second).String(),
			SpotifyItem: track,
		})
	}
	return items, nil
}

func PlaylistView(commands *commands.Commander, playlist spotify.SimplePlaylist) ([]list.Item, error) {
	items := []list.Item{}
	playlistItems, err := commands.Client().GetPlaylistItems(
		commands.Context,
		playlist.ID,
		spotify.Limit(50),
		spotify.Offset(0),
	)
	if err != nil {
		return nil, err
	}
	for _, item := range playlistItems.Items {
		items = append(items, mainItem{
			Name:     item.Track.Track.Name,
			Artist:   item.Track.Track.Artists[0],
			Duration: item.Track.Track.TimeDuration().Round(time.Second).String(),
			ID:       item.Track.Track.ID,
			Desc: item.Track.Track.Artists[0].Name + " - " + item.Track.Track.TimeDuration().
				Round(time.Second).
				String(),
			SpotifyItem: item,
		})
	}
	return items, nil
}

func ArtistsView(commands *commands.Commander) ([]list.Item, error) {
	items := []list.Item{}
	artists, err := commands.Client().CurrentUsersFollowedArtists(commands.Context, spotify.Limit(50), spotify.Offset(0))
	if err != nil {
		return nil, err
	}
	for _, artist := range artists.Artists {
		items = append(items, mainItem{
			Name:        artist.Name,
			ID:          artist.ID,
			Desc:        fmt.Sprintf("%d followers", artist.Followers.Count),
			SpotifyItem: artist.SimpleArtist,
		})
	}
	return items, nil
}

func SearchArtistsView(
	commands *commands.Commander,
	artists *spotify.FullArtistPage,
) ([]list.Item, error) {
	items := []list.Item{}
	for _, artist := range artists.Artists {
		items = append(items, mainItem{
			Name:        artist.Name,
			ID:          artist.ID,
			Desc:        fmt.Sprintf("%d followers", artist.Followers.Count),
			SpotifyItem: artist.SimpleArtist,
		})
	}
	return items, nil
}

func SearchView(commands *commands.Commander, search string) ([]list.Item, *SearchResults, error) {
	items := []list.Item{}

	result, err := commands.Search(search, 1)
	if err != nil {
		return nil, nil, err
	}
	items = append(
		items,
		mainItem{Name: "Tracks", Desc: "Search results", SpotifyItem: result.Tracks},
		mainItem{Name: "Albums", Desc: "Search results", SpotifyItem: result.Albums},
		mainItem{Name: "Artists", Desc: "Search results", SpotifyItem: result.Artists},
		mainItem{Name: "Playlists", Desc: "Search results", SpotifyItem: result.Playlists},
	)
	results := &SearchResults{
		Tracks:    result.Tracks,
		Playlists: result.Playlists,
		Albums:    result.Albums,
		Artists:   result.Artists,
	}
	return items, results, nil
}

func AlbumsView(commands *commands.Commander) ([]list.Item, error) {
	items := []list.Item{}
	albums, err := commands.Client().CurrentUsersAlbums(commands.Context, spotify.Limit(50), spotify.Offset((page-1)*50))
	if err != nil {
		return nil, err
	}
	for _, album := range albums.Albums {
		items = append(items, mainItem{
			Name: album.Name,
			ID:   album.ID,
			Desc: fmt.Sprintf(
				"%s by %s, %d tracks, released %d",
				album.AlbumType,
				album.Artists[0].Name,
				album.Tracks.Total,
				album.ReleaseDateTime().Year(),
			),
			SpotifyItem: album.SimpleAlbum,
		})
	}
	return items, nil
}

func SearchPlaylistsView(commands *commands.Commander, playlists *spotify.SimplePlaylistPage) ([]list.Item, error) {
	items := []list.Item{}
	for _, playlist := range playlists.Playlists {
		items = append(items, mainItem{
			Name:        playlist.Name,
			Desc:        stripHTMLRegex(playlist.Description),
			SpotifyItem: playlist,
		})
	}
	return items, nil
}

func SearchAlbumsView(commands *commands.Commander, albums *spotify.SimpleAlbumPage) ([]list.Item, error) {
	items := []list.Item{}
	for _, album := range albums.Albums {
		items = append(items, mainItem{
			Name: album.Name,
			ID:   album.ID,
			Desc: fmt.Sprintf(
				"%s by %s, released %d",
				album.AlbumType,
				album.Artists[0].Name,
				album.ReleaseDateTime().Year(),
			),
			SpotifyItem: album,
		})
	}
	return items, nil
}

func ArtistAlbumsView(album spotify.ID, commands *commands.Commander) ([]list.Item, error) {
	items := []list.Item{}
	albums, err := commands.ArtistAlbums(album, 1)
	if err != nil {
		return nil, err
	}
	for _, album := range albums.Albums {
		items = append(items, mainItem{
			Name:        album.Name,
			ID:          album.ID,
			Desc:        fmt.Sprintf("%s by %s", album.AlbumType, album.Artists[0].Name),
			SpotifyItem: album,
		})
	}
	return items, err
}

func AlbumTracksView(album spotify.ID, commands *commands.Commander) ([]list.Item, error) {
	items := []list.Item{}
	tracks, err := commands.AlbumTracks(album, 1)
	if err != nil {
		return nil, err
	}
	for _, track := range tracks.Tracks {
		items = append(items, mainItem{
			Name:        track.Name,
			Artist:      track.Artists[0],
			Duration:    track.TimeDuration().Round(time.Second).String(),
			ID:          track.ID,
			SpotifyItem: track,
			Desc:        track.Artists[0].Name + " - " + track.TimeDuration().Round(time.Second).String(),
		})
	}
	return items, err
}

func SearchTracksView(tracks *spotify.FullTrackPage) ([]list.Item, error) {
	items := []list.Item{}
	for _, track := range tracks.Tracks {
		items = append(items, mainItem{
			Name:        track.Name,
			Artist:      track.Artists[0],
			Duration:    track.TimeDuration().Round(time.Second).String(),
			ID:          track.ID,
			SpotifyItem: track,
			Desc:        track.Artists[0].Name + " - " + track.TimeDuration().Round(time.Second).String(),
		})
	}
	return items, nil
}

func SavedTracksView(commands *commands.Commander) ([]list.Item, error) {
	items := []list.Item{}
	tracks, err := commands.Client().CurrentUsersTracks(commands.Context, spotify.Limit(50), spotify.Offset((page-1)*50))
	if err != nil {
		return nil, err
	}
	for _, track := range tracks.Tracks {
		items = append(items, mainItem{
			Name:        track.Name,
			Artist:      track.Artists[0],
			Duration:    track.TimeDuration().Round(time.Second).String(),
			ID:          track.ID,
			SpotifyItem: track,
			Desc:        track.Artists[0].Name + " - " + track.TimeDuration().Round(time.Second).String(),
		})
	}
	return items, err
}

func MainView(c *commands.Commander) ([]list.Item, error) {
	c.Log.Debug("SWITCHING TO MAIN VIEW")
	wg := errgroup.Group{}
	var savedItems *spotify.SavedTrackPage
	var playlists *spotify.SimplePlaylistPage
	var artists *spotify.FullArtistCursorPage
	var albums *spotify.SavedAlbumPage

	wg.Go(func() (err error) {
		savedItems, err = c.Client().CurrentUsersTracks(c.Context, spotify.Limit(50), spotify.Offset(0))
		return
	})

	wg.Go(func() (err error) {
		playlists, err = c.Client().CurrentUsersPlaylists(c.Context, spotify.Limit(50), spotify.Offset(0))
		return
	})

	wg.Go(func() (err error) {
		artists, err = c.Client().CurrentUsersFollowedArtists(c.Context, spotify.Limit(50), spotify.Offset(0))
		return
	})

	wg.Go(func() (err error) {
		albums, err = c.Client().CurrentUsersAlbums(c.Context, spotify.Limit(50), spotify.Offset(0))
		return
	})

	err := wg.Wait()
	if err != nil {
		return nil, err
	}

	items := []list.Item{}
	if savedItems != nil && savedItems.Total != 0 {
		items = append(items, mainItem{
			Name:        "Saved Tracks",
			Desc:        fmt.Sprintf("%d saved songs", savedItems.Total),
			SpotifyItem: savedItems,
		})
	}
	if albums != nil && albums.Total != 0 {
		items = append(items, mainItem{
			Name:        "Albums",
			Desc:        fmt.Sprintf("%d albums", albums.Total),
			SpotifyItem: albums,
		})
	}
	if artists != nil && artists.Total != 0 {
		items = append(items, mainItem{
			Name:        "Artists",
			Desc:        fmt.Sprintf("%d artists", artists.Total),
			SpotifyItem: artists,
		})
	}
	items = append(items, mainItem{
		Name:        "Queue",
		Desc:        "Your Current Queue",
		SpotifyItem: spotify.Queue{},
	})
	if playlists != nil && playlists.Total != 0 {
		for _, playlist := range playlists.Playlists {
			items = append(items, mainItem{
				Name:        playlist.Name,
				Desc:        stripHTMLRegex(playlist.Description),
				SpotifyItem: playlist,
			})
		}
	}
	return items, nil
}

func stripHTMLRegex(s string) string {
	r := regexp.MustCompile(regex)
	return r.ReplaceAllString(s, "")
}
