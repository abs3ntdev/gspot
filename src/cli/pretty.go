package cli

import (
	"fmt"
	"io"
	"strings"
	"time"

	gspotv1 "github.com/abs3ntdev/gspot/gen/gspot/v1"
)

func prettyNowPlaying(w io.Writer, resp *gspotv1.NowPlayingResponse) error {
	icon := "▶"
	if !resp.IsPlaying {
		icon = "⏸"
	}
	if resp.Track == nil {
		fmt.Fprintf(w, "%s Nothing playing\n", icon)
		return nil
	}
	progress := formatDuration(resp.ProgressMs)
	duration := formatDuration(resp.Track.DurationMs)
	fmt.Fprintf(w, "%s %s - %s [%s/%s]\n", icon, resp.Track.Name, resp.Track.Artist, progress, duration)
	return nil
}

func prettyStatus(w io.Writer, resp *gspotv1.StatusResponse) error {
	if resp.State == nil {
		fmt.Fprintln(w, "No active player")
		return nil
	}
	s := resp.State
	icon := "▶"
	if !s.IsPlaying {
		icon = "⏸"
	}
	if s.Track != nil {
		fmt.Fprintf(w, "%s %s - %s (%s)\n", icon, s.Track.Name, s.Track.Artist, s.Track.Album)
	}
	progress := formatDuration(s.ProgressMs)
	duration := int32(0)
	if s.Track != nil {
		duration = s.Track.DurationMs
	}
	fmt.Fprintf(w, "  Progress: %s / %s\n", progress, formatDuration(duration))
	fmt.Fprintf(w, "  Shuffle: %v  Repeat: %s\n", s.ShuffleState, s.RepeatState)
	if s.Device != nil {
		fmt.Fprintf(w, "  Device: %s (%s) Volume: %d%%\n", s.Device.Name, s.Device.Type, s.Device.VolumePercent)
	}
	return nil
}

func prettyDevices(w io.Writer, resp *gspotv1.ListDevicesResponse) error {
	if len(resp.Devices) == 0 {
		fmt.Fprintln(w, "No devices found")
		return nil
	}
	for _, d := range resp.Devices {
		active := " "
		if d.IsActive {
			active = "*"
		}
		fmt.Fprintf(w, " %s %-30s %-10s Volume: %d%%\n", active, d.Name, d.Type, d.VolumePercent)
	}
	return nil
}

func prettyLink(w io.Writer, resp *gspotv1.GetLinkResponse) error {
	fmt.Fprintln(w, resp.Url)
	return nil
}

func prettyLinkContext(w io.Writer, resp *gspotv1.GetLinkContextResponse) error {
	fmt.Fprintln(w, resp.Url)
	return nil
}

func prettyYoutubeLink(w io.Writer, resp *gspotv1.GetYoutubeLinkResponse) error {
	fmt.Fprintln(w, resp.Url)
	return nil
}

func prettyVolume(w io.Writer, resp *gspotv1.ChangeVolumeResponse) error {
	fmt.Fprintf(w, "Volume: %d%%\n", resp.VolumePercent)
	return nil
}

func prettyRepeat(w io.Writer, resp *gspotv1.RepeatResponse) error {
	fmt.Fprintf(w, "Repeat: %s\n", resp.State)
	return nil
}

func prettyShuffle(w io.Writer, resp *gspotv1.ShuffleResponse) error {
	state := "off"
	if resp.State {
		state = "on"
	}
	fmt.Fprintf(w, "Shuffle: %s\n", state)
	return nil
}

func prettySetVolume(w io.Writer, resp *gspotv1.SetVolumeResponse) error {
	fmt.Fprintf(w, "Volume: %d%%\n", resp.VolumePercent)
	return nil
}

func prettyDevice(w io.Writer, resp *gspotv1.SetDeviceResponse) error {
	if resp.Device == nil {
		return nil
	}
	fmt.Fprintf(w, "Active device: %s (%s)\n", resp.Device.Name, resp.Device.Type)
	return nil
}

func prettyCover(w io.Writer, resp *gspotv1.DownloadCoverResponse) error {
	fmt.Fprintf(w, "Cover saved to %s\n", resp.Path)
	return nil
}

func prettyPlaylists(w io.Writer, resp *gspotv1.ListPlaylistsResponse) error {
	if len(resp.Playlists) == 0 {
		fmt.Fprintln(w, "No playlists found")
		return nil
	}
	for _, p := range resp.Playlists {
		fmt.Fprintf(w, "  %-40s %3d tracks  %s\n", p.Name, p.TrackCount, p.Id)
	}
	fmt.Fprintf(w, "\n%d playlists total\n", resp.Total)
	return nil
}

func prettyPlaylist(w io.Writer, resp *gspotv1.GetPlaylistResponse) error {
	if resp.Playlist == nil {
		fmt.Fprintln(w, "Playlist not found")
		return nil
	}
	fmt.Fprintf(w, "%s (%d tracks) by %s\n", resp.Playlist.Name, resp.Playlist.TrackCount, resp.Playlist.Owner)
	fmt.Fprintln(w, strings.Repeat("─", 60))
	for i, t := range resp.Tracks {
		duration := formatDuration(t.DurationMs)
		fmt.Fprintf(w, "  %3d. %-35s %-20s %s\n", i+1, t.Name, t.Artist, duration)
	}
	return nil
}

// prettyEmpty is for commands with no meaningful response data.
func prettyEmpty[T any](w io.Writer, resp T) error {
	return nil
}

func formatDuration(ms int32) string {
	d := time.Duration(ms) * time.Millisecond
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	if minutes >= 60 {
		hours := minutes / 60
		minutes = minutes % 60
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}
