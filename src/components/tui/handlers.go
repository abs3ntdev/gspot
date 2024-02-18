package tui

import (
	"github.com/zmb3/spotify/v2"

	"git.asdf.cafe/abs3nt/gospt-ng/src/components/commands"
)

func HandlePlayWithContext(commands *commands.Commander, uri *spotify.URI, pos *int) {
	err := commands.PlaySongInPlaylist(uri, pos)
	if err != nil {
		return
	}
}

func HandleRadio(commands *commands.Commander, song spotify.SimpleTrack) {
	err := commands.RadioGivenSong(song, 0)
	if err != nil {
		return
	}
}

func HandleAlbumRadio(commands *commands.Commander, album spotify.SimpleAlbum) {
	err := commands.RadioFromAlbum(album)
	if err != nil {
		return
	}
}

func HandleSeek(commands *commands.Commander, fwd bool) {
	err := commands.Seek(fwd)
	if err != nil {
		return
	}
}

func HandleVolume(commands *commands.Commander, up bool) {
	vol := 10
	if !up {
		vol = -10
	}
	err := commands.ChangeVolume(vol)
	if err != nil {
		return
	}
}

func HandleArtistRadio(commands *commands.Commander, artist spotify.SimpleArtist) {
	err := commands.RadioGivenArtist(artist)
	if err != nil {
		return
	}
}

func HandleAlbumArtist(commands *commands.Commander, artist spotify.SimpleArtist) {
	err := commands.RadioGivenArtist(artist)
	if err != nil {
		return
	}
}

func HandlePlaylistRadio(commands *commands.Commander, playlist spotify.SimplePlaylist) {
	err := commands.RadioFromPlaylist(playlist)
	if err != nil {
		return
	}
}

func HandleLibraryRadio(commands *commands.Commander) {
	err := commands.RadioFromSavedTracks()
	if err != nil {
		return
	}
}

func HandlePlayLikedSong(commands *commands.Commander, position int) {
	err := commands.PlayLikedSongs(position)
	if err != nil {
		return
	}
}

func HandlePlayTrack(commands *commands.Commander, track spotify.ID) {
	err := commands.QueueSong(track)
	if err != nil {
		return
	}
	err = commands.Next()
	if err != nil {
		return
	}
}

func HandleNextInQueue(commands *commands.Commander, amt int) {
	err := commands.Next()
	if err != nil {
		return
	}
}

func HandleQueueItem(commands *commands.Commander, item spotify.ID) {
	err := commands.QueueSong(item)
	if err != nil {
		return
	}
}

func HandleDeleteTrackFromPlaylist(commands *commands.Commander, item, playlist spotify.ID) {
	err := commands.DeleteTracksFromPlaylist([]spotify.ID{item}, playlist)
	if err != nil {
		return
	}
}

func HandleSetDevice(commands *commands.Commander, player spotify.PlayerDevice) {
	err := commands.SetDevice(player.ID)
	if err != nil {
		return
	}
}
