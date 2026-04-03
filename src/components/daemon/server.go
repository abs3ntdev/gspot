package daemon

import (
	"context"
	"path/filepath"

	"connectrpc.com/connect"
	"github.com/zmb3/spotify/v2"

	gspotv1 "github.com/abs3ntdev/gspot/gen/gspot/v1"
	"github.com/abs3ntdev/gspot/src/components/commands"
)

// Server implements the GspotService ConnectRPC handler.
type Server struct {
	commander *commands.Commander
}

func NewServer(c *commands.Commander) *Server {
	return &Server{commander: c}
}

// Playback

func (s *Server) Play(ctx context.Context, req *connect.Request[gspotv1.PlayRequest]) (*connect.Response[gspotv1.PlayResponse], error) {
	if err := s.commander.Play(); err != nil {
		return nil, err
	}
	return connect.NewResponse(&gspotv1.PlayResponse{}), nil
}

func (s *Server) PlayURL(ctx context.Context, req *connect.Request[gspotv1.PlayURLRequest]) (*connect.Response[gspotv1.PlayURLResponse], error) {
	if err := s.commander.PlayURL(req.Msg.Url); err != nil {
		return nil, err
	}
	return connect.NewResponse(&gspotv1.PlayURLResponse{}), nil
}

func (s *Server) Pause(ctx context.Context, req *connect.Request[gspotv1.PauseRequest]) (*connect.Response[gspotv1.PauseResponse], error) {
	if err := s.commander.Pause(); err != nil {
		return nil, err
	}
	return connect.NewResponse(&gspotv1.PauseResponse{}), nil
}

func (s *Server) TogglePlay(ctx context.Context, req *connect.Request[gspotv1.TogglePlayRequest]) (*connect.Response[gspotv1.TogglePlayResponse], error) {
	if err := s.commander.TogglePlay(); err != nil {
		return nil, err
	}
	return connect.NewResponse(&gspotv1.TogglePlayResponse{}), nil
}

func (s *Server) Next(ctx context.Context, req *connect.Request[gspotv1.NextRequest]) (*connect.Response[gspotv1.NextResponse], error) {
	amount := int(req.Msg.Amount)
	if amount == 0 {
		amount = 1
	}
	if err := s.commander.Next(amount, false); err != nil {
		return nil, err
	}
	return connect.NewResponse(&gspotv1.NextResponse{}), nil
}

func (s *Server) Previous(ctx context.Context, req *connect.Request[gspotv1.PreviousRequest]) (*connect.Response[gspotv1.PreviousResponse], error) {
	if err := s.commander.Previous(); err != nil {
		return nil, err
	}
	return connect.NewResponse(&gspotv1.PreviousResponse{}), nil
}

func (s *Server) Seek(ctx context.Context, req *connect.Request[gspotv1.SeekRequest]) (*connect.Response[gspotv1.SeekResponse], error) {
	if err := s.commander.Seek(req.Msg.Forward); err != nil {
		return nil, err
	}
	return connect.NewResponse(&gspotv1.SeekResponse{}), nil
}

func (s *Server) SetPosition(ctx context.Context, req *connect.Request[gspotv1.SetPositionRequest]) (*connect.Response[gspotv1.SetPositionResponse], error) {
	if err := s.commander.SetPosition(int(req.Msg.PositionMs)); err != nil {
		return nil, err
	}
	return connect.NewResponse(&gspotv1.SetPositionResponse{}), nil
}

func (s *Server) ChangeVolume(ctx context.Context, req *connect.Request[gspotv1.ChangeVolumeRequest]) (*connect.Response[gspotv1.ChangeVolumeResponse], error) {
	if err := s.commander.ChangeVolume(int(req.Msg.Amount)); err != nil {
		return nil, err
	}
	// Read back current volume.
	state, err := s.commander.Client().PlayerState(s.commander.Context)
	vol := int32(0)
	if err == nil && state != nil {
		vol = int32(state.Device.Volume)
	}
	return connect.NewResponse(&gspotv1.ChangeVolumeResponse{VolumePercent: vol}), nil
}

func (s *Server) Mute(ctx context.Context, req *connect.Request[gspotv1.MuteRequest]) (*connect.Response[gspotv1.MuteResponse], error) {
	if err := s.commander.Mute(); err != nil {
		return nil, err
	}
	return connect.NewResponse(&gspotv1.MuteResponse{}), nil
}

func (s *Server) UnMute(ctx context.Context, req *connect.Request[gspotv1.UnMuteRequest]) (*connect.Response[gspotv1.UnMuteResponse], error) {
	if err := s.commander.UnMute(); err != nil {
		return nil, err
	}
	return connect.NewResponse(&gspotv1.UnMuteResponse{}), nil
}

func (s *Server) ToggleMute(ctx context.Context, req *connect.Request[gspotv1.ToggleMuteRequest]) (*connect.Response[gspotv1.ToggleMuteResponse], error) {
	if err := s.commander.ToggleMute(); err != nil {
		return nil, err
	}
	return connect.NewResponse(&gspotv1.ToggleMuteResponse{}), nil
}

func (s *Server) Repeat(ctx context.Context, req *connect.Request[gspotv1.RepeatRequest]) (*connect.Response[gspotv1.RepeatResponse], error) {
	if err := s.commander.Repeat(); err != nil {
		return nil, err
	}
	// Read back repeat state.
	state, err := s.commander.Client().PlayerState(s.commander.Context)
	repeatState := ""
	if err == nil && state != nil {
		repeatState = state.RepeatState
	}
	return connect.NewResponse(&gspotv1.RepeatResponse{State: repeatState}), nil
}

func (s *Server) Shuffle(ctx context.Context, req *connect.Request[gspotv1.ShuffleRequest]) (*connect.Response[gspotv1.ShuffleResponse], error) {
	if err := s.commander.Shuffle(); err != nil {
		return nil, err
	}
	// Read back shuffle state.
	state, err := s.commander.Client().PlayerState(s.commander.Context)
	shuffleState := false
	if err == nil && state != nil {
		shuffleState = state.ShuffleState
	}
	return connect.NewResponse(&gspotv1.ShuffleResponse{State: shuffleState}), nil
}

func (s *Server) SetDevice(ctx context.Context, req *connect.Request[gspotv1.SetDeviceRequest]) (*connect.Response[gspotv1.SetDeviceResponse], error) {
	if err := s.commander.SetDevice(spotify.ID(req.Msg.DeviceId)); err != nil {
		return nil, err
	}
	// Try to find the device info.
	devices, err := s.commander.Client().PlayerDevices(s.commander.Context)
	resp := &gspotv1.SetDeviceResponse{}
	if err == nil {
		for _, d := range devices {
			if string(d.ID) == req.Msg.DeviceId {
				resp.Device = deviceToProto(d)
				break
			}
		}
	}
	return connect.NewResponse(resp), nil
}

func (s *Server) QueueSong(ctx context.Context, req *connect.Request[gspotv1.QueueSongRequest]) (*connect.Response[gspotv1.QueueSongResponse], error) {
	if err := s.commander.QueueSong(spotify.ID(req.Msg.TrackId)); err != nil {
		return nil, err
	}
	return connect.NewResponse(&gspotv1.QueueSongResponse{}), nil
}

// Library

func (s *Server) Like(ctx context.Context, req *connect.Request[gspotv1.LikeRequest]) (*connect.Response[gspotv1.LikeResponse], error) {
	if err := s.commander.Like(); err != nil {
		return nil, err
	}
	return connect.NewResponse(&gspotv1.LikeResponse{}), nil
}

func (s *Server) UnLike(ctx context.Context, req *connect.Request[gspotv1.UnLikeRequest]) (*connect.Response[gspotv1.UnLikeResponse], error) {
	if err := s.commander.UnLike(); err != nil {
		return nil, err
	}
	return connect.NewResponse(&gspotv1.UnLikeResponse{}), nil
}

// Info / queries

func (s *Server) NowPlaying(ctx context.Context, req *connect.Request[gspotv1.NowPlayingRequest]) (*connect.Response[gspotv1.NowPlayingResponse], error) {
	current, err := s.commander.Client().PlayerCurrentlyPlaying(s.commander.Context)
	if err != nil {
		return nil, err
	}
	resp := &gspotv1.NowPlayingResponse{}
	if current != nil {
		resp.IsPlaying = current.Playing
		resp.ProgressMs = int32(current.Progress)
		if current.Item != nil {
			resp.Track = &gspotv1.Track{
				Id:         string(current.Item.ID),
				Name:       current.Item.Name,
				Album:      current.Item.Album.Name,
				DurationMs: int32(current.Item.Duration),
			}
			if len(current.Item.Artists) > 0 {
				resp.Track.Artist = current.Item.Artists[0].Name
			}
			if urls := current.Item.ExternalURLs; urls != nil {
				resp.Track.SpotifyUrl = urls["spotify"]
			}
		}
	}
	return connect.NewResponse(resp), nil
}

func (s *Server) Status(ctx context.Context, req *connect.Request[gspotv1.StatusRequest]) (*connect.Response[gspotv1.StatusResponse], error) {
	state, err := s.commander.Client().PlayerState(s.commander.Context)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&gspotv1.StatusResponse{
		State: playerStateToProto(state),
	}), nil
}

func (s *Server) ListDevices(ctx context.Context, req *connect.Request[gspotv1.ListDevicesRequest]) (*connect.Response[gspotv1.ListDevicesResponse], error) {
	devices, err := s.commander.Client().PlayerDevices(s.commander.Context)
	if err != nil {
		return nil, err
	}
	resp := &gspotv1.ListDevicesResponse{}
	for _, d := range devices {
		resp.Devices = append(resp.Devices, deviceToProto(d))
	}
	return connect.NewResponse(resp), nil
}

func (s *Server) GetLink(ctx context.Context, req *connect.Request[gspotv1.GetLinkRequest]) (*connect.Response[gspotv1.GetLinkResponse], error) {
	link, err := s.commander.PrintLink()
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&gspotv1.GetLinkResponse{Url: link}), nil
}

func (s *Server) GetLinkContext(ctx context.Context, req *connect.Request[gspotv1.GetLinkContextRequest]) (*connect.Response[gspotv1.GetLinkContextResponse], error) {
	link, err := s.commander.PrintLinkContext()
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&gspotv1.GetLinkContextResponse{Url: link}), nil
}

func (s *Server) GetYoutubeLink(ctx context.Context, req *connect.Request[gspotv1.GetYoutubeLinkRequest]) (*connect.Response[gspotv1.GetYoutubeLinkResponse], error) {
	link, err := s.commander.PrintYoutubeLink()
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&gspotv1.GetYoutubeLinkResponse{Url: link}), nil
}

func (s *Server) DownloadCover(ctx context.Context, req *connect.Request[gspotv1.DownloadCoverRequest]) (*connect.Response[gspotv1.DownloadCoverResponse], error) {
	path := req.Msg.Path
	if path == "" {
		path = "cover.png"
	}
	path = filepath.Clean(path)
	if err := s.commander.DownloadCover(path); err != nil {
		return nil, err
	}
	return connect.NewResponse(&gspotv1.DownloadCoverResponse{Path: path}), nil
}

// Playlists

func (s *Server) ListPlaylists(ctx context.Context, req *connect.Request[gspotv1.ListPlaylistsRequest]) (*connect.Response[gspotv1.ListPlaylistsResponse], error) {
	limit := int(req.Msg.Limit)
	if limit <= 0 {
		limit = 50
	}
	offset := int(req.Msg.Offset)

	playlists, err := s.commander.Client().CurrentUsersPlaylists(s.commander.Context, spotify.Limit(limit), spotify.Offset(offset))
	if err != nil {
		return nil, err
	}

	resp := &gspotv1.ListPlaylistsResponse{
		Total: int32(playlists.Total),
	}
	for _, p := range playlists.Playlists {
		resp.Playlists = append(resp.Playlists, playlistToProto(p))
	}
	return connect.NewResponse(resp), nil
}

func (s *Server) GetPlaylist(ctx context.Context, req *connect.Request[gspotv1.GetPlaylistRequest]) (*connect.Response[gspotv1.GetPlaylistResponse], error) {
	playlist, err := s.commander.Client().GetPlaylist(s.commander.Context, spotify.ID(req.Msg.PlaylistId))
	if err != nil {
		return nil, err
	}

	resp := &gspotv1.GetPlaylistResponse{
		Playlist: &gspotv1.Playlist{
			Id:         string(playlist.ID),
			Name:       playlist.Name,
			Owner:      playlist.Owner.DisplayName,
			TrackCount: int32(playlist.Tracks.Total),
			IsPublic:   playlist.IsPublic,
		},
	}
	if urls := playlist.ExternalURLs; urls != nil {
		resp.Playlist.SpotifyUrl = urls["spotify"]
	}

	for _, item := range playlist.Tracks.Tracks {
		t := item.Track
		track := &gspotv1.Track{
			Id:         string(t.ID),
			Name:       t.Name,
			Album:      t.Album.Name,
			DurationMs: int32(t.Duration),
		}
		if len(t.Artists) > 0 {
			track.Artist = t.Artists[0].Name
		}
		if urls := t.ExternalURLs; urls != nil {
			track.SpotifyUrl = urls["spotify"]
		}
		resp.Tracks = append(resp.Tracks, track)
	}
	return connect.NewResponse(resp), nil
}

func (s *Server) PlayPlaylist(ctx context.Context, req *connect.Request[gspotv1.PlayPlaylistRequest]) (*connect.Response[gspotv1.PlayPlaylistResponse], error) {
	uri := spotify.URI("spotify:playlist:" + req.Msg.PlaylistId)
	opts := &spotify.PlayOptions{
		PlaybackContext: &uri,
	}
	if req.Msg.Offset > 0 {
		offset := int(req.Msg.Offset)
		opts.PlaybackOffset = &spotify.PlaybackOffset{Position: &offset}
	}
	err := s.commander.Client().PlayOpt(s.commander.Context, opts)
	if err != nil {
		if commands.IsNoActiveError(err) {
			deviceID, err := s.commander.ActivateDevice()
			if err != nil {
				return nil, err
			}
			opts.DeviceID = &deviceID
			err = s.commander.Client().PlayOpt(s.commander.Context, opts)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return connect.NewResponse(&gspotv1.PlayPlaylistResponse{}), nil
}

// Helpers

func playlistToProto(p spotify.SimplePlaylist) *gspotv1.Playlist {
	pl := &gspotv1.Playlist{
		Id:         string(p.ID),
		Name:       p.Name,
		Owner:      p.Owner.DisplayName,
		TrackCount: int32(p.Tracks.Total),
		IsPublic:   p.IsPublic,
	}
	if urls := p.ExternalURLs; urls != nil {
		pl.SpotifyUrl = urls["spotify"]
	}
	return pl
}

func playerStateToProto(state *spotify.PlayerState) *gspotv1.PlayerState {
	if state == nil {
		return nil
	}
	ps := &gspotv1.PlayerState{
		IsPlaying:    state.Playing,
		ProgressMs:   int32(state.Progress),
		RepeatState:  state.RepeatState,
		ShuffleState: state.ShuffleState,
		Device:       deviceToProto(state.Device),
	}
	if state.Item != nil {
		ps.Track = &gspotv1.Track{
			Id:         string(state.Item.ID),
			Name:       state.Item.Name,
			Album:      state.Item.Album.Name,
			DurationMs: int32(state.Item.Duration),
		}
		if len(state.Item.Artists) > 0 {
			ps.Track.Artist = state.Item.Artists[0].Name
		}
		if urls := state.Item.ExternalURLs; urls != nil {
			ps.Track.SpotifyUrl = urls["spotify"]
		}
	}
	return ps
}

func deviceToProto(d spotify.PlayerDevice) *gspotv1.Device {
	return &gspotv1.Device{
		Id:            string(d.ID),
		Name:          d.Name,
		Type:          d.Type,
		VolumePercent: int32(d.Volume),
		IsActive:      d.Active,
	}
}
