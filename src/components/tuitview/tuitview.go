package tuitview

import (
	"sync/atomic"

	"github.com/abs3ntdev/gspot/src/components/commands"
	"github.com/rivo/tview"
	"github.com/zmb3/spotify/v2"
)

var (
	tracksLoading    = atomic.Bool{}
	playlistsLoading = atomic.Bool{}
	tracksPage       = 1
	playlistsPage    = 1
)

func TuitView(cmd *commands.Commander) error {
	tracksLoading.Store(false)
	playlistsLoading.Store(false)
	playlistsList := tview.NewList().ShowSecondaryText(false)
	playlistsList.SetBorder(true).SetTitle("Playlists")
	savedTracksList := tview.NewList().ShowSecondaryText(false)
	savedTracksList.SetWrapAround(false)
	savedTracksList.SetBorder(true).SetTitle("Tracks")
	savedTracksList.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		go cmd.PlayLikedSongs(index)
	})
	flex := tview.NewFlex().AddItem(playlistsList, 0, 1, false).AddItem(savedTracksList, 0, 2, true)
	playlists, err := cmd.Playlists(1)
	if err != nil {
		return err
	}
	for _, playlist := range playlists.Playlists {
		playlistsList.AddItem(playlist.Name, "", 0, func() {
			playlistTracksList := tview.NewList().ShowSecondaryText(false)
			playlistTracksList.SetBorder(true).SetTitle(playlist.Name)
			playlistTracksList.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
				go cmd.PlaySongInPlaylist((*spotify.URI)(&secondaryText), &index)
			})
			tracks, err := cmd.PlaylistTracks(playlist.ID, 1)
			if err != nil {
				return
			}
			for _, track := range tracks.Items {
				playlistTracksList.AddItem(track.Track.Track.Name+" - "+track.Track.Track.Artists[0].Name, string(playlist.URI), 0, nil)
			}
			flex.Clear()
			flex.AddItem(playlistsList, 0, 1, false)
			flex.AddItem(playlistTracksList, 0, 2, false)
		})
	}
	tracks, err := cmd.TrackList(1)
	if err != nil {
		return err
	}
	for _, track := range tracks.Tracks {
		savedTracksList.AddItem(track.Name+" - "+track.Artists[0].Name, "", 0, nil)
	}
	playlistsList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if playlistsList.GetItemCount()%50 != 0 {
			return
		}
		if playlistsList.GetItemCount()-index < 40 {
			go func() {
				if playlistsLoading.Load() {
					return
				}
				playlistsLoading.Store(true)
				defer playlistsLoading.Store(false)
				playlistsPage++
				newPlaylists, _ := cmd.Playlists(playlistsPage)
				for _, playlist := range newPlaylists.Playlists {
					savedTracksList.AddItem(playlist.Name, "", 0, nil)
				}
			}()
		}
	})
	savedTracksList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if savedTracksList.GetItemCount()%50 != 0 {
			return
		}
		if savedTracksList.GetItemCount()-index < 40 {
			go func() {
				if tracksLoading.Load() {
					return
				}
				tracksLoading.Store(true)
				defer tracksLoading.Store(false)
				tracksPage++
				tracks, _ := cmd.TrackList(tracksPage)
				for _, track := range tracks.Tracks {
					savedTracksList.AddItem(track.Name+" - "+track.Artists[0].Name, "", 0, nil)
				}
			}()
		}
	})
	if err := tview.NewApplication().EnableMouse(true).SetRoot(flex, true).Run(); err != nil {
		return err
	}
	return nil
}
