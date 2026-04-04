package librespot

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	golibrespot "github.com/devgianlu/go-librespot"
	"github.com/devgianlu/go-librespot/ap"
	"github.com/devgianlu/go-librespot/apresolve"
	"github.com/devgianlu/go-librespot/dealer"
	"github.com/devgianlu/go-librespot/player"
	connectpb "github.com/devgianlu/go-librespot/proto/spotify/connectstate"
	devicespb "github.com/devgianlu/go-librespot/proto/spotify/connectstate/devices"
	"github.com/devgianlu/go-librespot/session"
	"github.com/devgianlu/go-librespot/tracks"
	"github.com/devgianlu/go-librespot/zeroconf"
	"go.uber.org/fx"
	"google.golang.org/protobuf/proto"

	gspotv1 "github.com/abs3ntdev/gspot/gen/gspot/v1"
	"github.com/abs3ntdev/gspot/src/config"
)

// Player is the librespot-based Spotify Connect player.
// It manages a session, audio player, and Spotify Connect state.
type Player struct {
	// Injected
	conf *config.Config
	log  golibrespot.Logger

	// Librespot internals
	sess            *session.Session
	audioPlayer     *player.Player
	state           *state
	appState        *golibrespot.AppState
	httpClient      *http.Client
	resolver        *apresolve.ApResolver
	deviceId        string
	deviceType      devicespb.DeviceType
	spotConnId      string
	countryCode     *string
	primaryStream   *player.Stream
	secondaryStream *player.Stream
	prefetchTimer   *time.Timer
	volumeUpdate    chan float32

	// Event broadcasting
	events        chan *gspotv1.SubscribeResponse
	subscribersMu sync.RWMutex
	subscribers   []chan *gspotv1.SubscribeResponse

	// Public state
	mu     sync.RWMutex
	active bool

	// Lifecycle
	cancel context.CancelFunc
}

// NewPlayer creates a new librespot Player. Returns nil if librespot is not enabled in config.
// Registers fx lifecycle hooks for start/stop.
func NewPlayer(lc fx.Lifecycle, conf *config.Config, log golibrespot.Logger) *Player {
	if !conf.Librespot.Enabled {
		return nil
	}

	p := &Player{
		conf:         conf,
		log:          log,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		countryCode:  new(string),
		volumeUpdate: make(chan float32, 1),
		events:       make(chan *gspotv1.SubscribeResponse, 64),
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return p.start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return p.stop()
		},
	})

	return p
}

// DeviceId returns the librespot device ID.
func (p *Player) DeviceId() string {
	if p == nil {
		return ""
	}
	return p.deviceId
}

// IsActive returns true if this device is the currently active Spotify device.
func (p *Player) IsActive() bool {
	if p == nil {
		return false
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.active
}

// GetState returns the current player state as a proto message.
func (p *Player) GetState() *gspotv1.PlayerState {
	if p == nil || p.state == nil {
		return nil
	}

	ps := &gspotv1.PlayerState{
		IsPlaying:    p.state.player.IsPlaying && !p.state.player.IsPaused,
		ProgressMs:   int32(p.state.trackPosition()),
		RepeatState:  repeatStateString(p.state.player.Options),
		ShuffleState: p.state.player.Options.ShufflingContext,
	}

	if p.state.device != nil {
		ps.Device = &gspotv1.Device{
			Id:            p.deviceId,
			Name:          p.conf.Librespot.DeviceName,
			Type:          p.deviceType.String(),
			VolumePercent: int32(p.apiVolume()),
			IsActive:      p.state.active,
		}
	}

	if p.primaryStream != nil && p.primaryStream.Media != nil {
		ps.Track = &gspotv1.Track{
			Name:       p.primaryStream.Media.Name(),
			DurationMs: p.primaryStream.Media.Duration(),
		}
		// Extract artist and album from media if available
		if t := p.primaryStream.Media.Track(); t != nil {
			if len(t.Artist) > 0 {
				ps.Track.Artist = t.Artist[0].GetName()
			}
			if t.Album != nil {
				ps.Track.Album = t.Album.GetName()
			}
		}
		if p.state.player.Track != nil {
			id, _ := golibrespot.SpotifyIdFromUri(p.state.player.Track.Uri)
			if id != nil {
				ps.Track.Id = id.String()
				ps.Track.SpotifyUrl = "https://open.spotify.com/track/" + id.String()
			}
		}
	}

	return ps
}

// Subscribe returns a channel that receives player events.
// Each subscriber gets their own bounded channel (64 events).
func (p *Player) Subscribe() <-chan *gspotv1.SubscribeResponse {
	ch := make(chan *gspotv1.SubscribeResponse, 64)
	p.subscribersMu.Lock()
	p.subscribers = append(p.subscribers, ch)
	p.subscribersMu.Unlock()
	return ch
}

// Play resumes playback on the librespot player.
func (p *Player) Play(ctx context.Context) error {
	return p.play(ctx)
}

// Pause pauses playback.
func (p *Player) Pause(ctx context.Context) error {
	return p.pause(ctx)
}

// Next skips to the next track.
func (p *Player) Next(ctx context.Context) error {
	return p.skipNext(ctx, nil)
}

// Previous skips to the previous track.
func (p *Player) Previous(ctx context.Context) error {
	return p.skipPrev(ctx, true)
}

// Seek seeks to a position in milliseconds.
func (p *Player) Seek(ctx context.Context, posMs int64) error {
	return p.seek(ctx, posMs)
}

// SetVolume sets the volume (0 - MaxStateVolume).
func (p *Player) SetVolume(ctx context.Context, vol uint32) error {
	p.updateVolume(vol * player.MaxStateVolume / p.conf.Librespot.VolumeSteps)
	return nil
}

// SetShuffle sets shuffle state.
func (p *Player) SetShuffle(ctx context.Context, enabled bool) error {
	p.setOptions(ctx, nil, nil, &enabled)
	return nil
}

// SetRepeat sets repeat state.
func (p *Player) SetRepeat(ctx context.Context, repeatState string) error {
	switch repeatState {
	case "context":
		t := true
		f := false
		p.setOptions(ctx, &t, &f, nil)
	case "track":
		t := true
		p.setOptions(ctx, &t, &t, nil)
	default: // "off"
		f := false
		p.setOptions(ctx, &f, &f, nil)
	}
	return nil
}

// LoadContext loads a Spotify context (playlist, album, etc.) for playback.
func (p *Player) LoadContext(ctx context.Context, uri string, offset int) error {
	spotCtx, err := p.sess.Spclient().ContextResolve(ctx, uri)
	if err != nil {
		return fmt.Errorf("failed resolving context: %w", err)
	}

	p.mu.Lock()
	p.active = true
	p.mu.Unlock()

	p.state.setActive(true)
	p.state.setPaused(false)
	p.state.player.Suppressions = &connectpb.Suppressions{}
	p.state.player.PlayOrigin = &connectpb.PlayOrigin{
		FeatureIdentifier: "gspot",
	}

	var skipTo skipToFunc
	if offset > 0 {
		idx := -1
		skipTo = func(_ *connectpb.ContextTrack) bool {
			idx++
			return idx == offset
		}
	}

	return p.loadContext(ctx, spotCtx, skipTo, false, true)
}

// start initializes the session and starts the event loop.
func (p *Player) start(ctx context.Context) error {
	ctx, p.cancel = context.WithCancel(ctx)

	// Load stored credentials
	p.appState = &golibrespot.AppState{}
	p.appState.SetLogger(p.log)
	configDir := config.ConfigDir()
	if err := p.appState.Read(configDir); err != nil {
		p.log.WithError(err).Warn("failed reading librespot state, starting fresh")
	}

	p.resolver = apresolve.NewApResolver(p.log, p.httpClient)

	// Resolve device ID
	if p.appState.DeviceId != "" {
		p.deviceId = p.appState.DeviceId
	} else {
		deviceIdBytes := make([]byte, 20)
		_, _ = rand.Read(deviceIdBytes)
		p.deviceId = hex.EncodeToString(deviceIdBytes)
		p.appState.DeviceId = p.deviceId
		_ = p.appState.Write()
	}

	p.deviceType = devicespb.DeviceType_COMPUTER

	// Check if we have stored credentials
	if len(p.appState.Credentials.Data) == 0 {
		p.log.Info("No librespot credentials found. Run 'gspot librespot auth' to authenticate.")
		return nil // Don't fail — daemon still works without librespot
	}

	// Create session with stored credentials
	var err error
	p.sess, err = session.NewSessionFromOptions(ctx, &session.Options{
		Log:        p.log,
		DeviceType: p.deviceType,
		DeviceId:   p.deviceId,
		Credentials: session.StoredCredentials{
			Username: p.appState.Credentials.Username,
			Data:     p.appState.Credentials.Data,
		},
		Resolver: p.resolver,
		Client:   p.httpClient,
		AppState: p.appState,
	})
	if err != nil {
		return fmt.Errorf("failed creating librespot session: %w", err)
	}

	p.prefetchTimer = time.NewTimer(math.MaxInt64)
	p.prefetchTimer.Stop()

	// Init state and create audio player
	p.initState()

	p.audioPlayer, err = player.NewPlayer(&player.Options{
		Spclient:             p.sess.Spclient(),
		AudioKey:             p.sess.AudioKey(),
		Events:               p.sess.Events(),
		Log:                  p.log,
		NormalisationEnabled: p.conf.Librespot.Normalise,
		CountryCode:          p.countryCode,
		AudioBackend:         p.conf.Librespot.AudioBackend,
		AudioDevice:          p.conf.Librespot.AudioDevice,
		VolumeUpdate:         p.volumeUpdate,
	})
	if err != nil {
		return fmt.Errorf("failed initializing audio player: %w", err)
	}

	// Set initial volume
	if lastVolume := p.appState.LastVolume; lastVolume != nil {
		p.updateVolume(*lastVolume)
	} else {
		p.updateVolume(p.conf.Librespot.InitialVol * player.MaxStateVolume / p.conf.Librespot.VolumeSteps)
	}

	// Start event loop
	go p.run(ctx)

	// Start zeroconf if enabled
	if p.conf.Librespot.Zeroconf {
		go p.startZeroconf(ctx)
	}

	p.log.Infof("librespot player started (device: %s, id: %s...)", p.conf.Librespot.DeviceName, p.deviceId[:8])
	return nil
}

// stop shuts down the librespot player.
func (p *Player) stop() error {
	if p.cancel != nil {
		p.cancel()
	}
	if p.audioPlayer != nil {
		p.audioPlayer.Close()
	}
	if p.sess != nil {
		p.sess.Close()
	}

	// Close all subscriber channels
	p.subscribersMu.Lock()
	for _, ch := range p.subscribers {
		close(ch)
	}
	p.subscribers = nil
	p.subscribersMu.Unlock()

	p.log.Info("librespot player stopped")
	return nil
}

// run is the main event loop — adapted from go-librespot's AppPlayer.Run().
func (p *Player) run(ctx context.Context) {
	if err := p.sess.Dealer().Connect(ctx); err != nil {
		p.log.WithError(err).Error("failed connecting to dealer")
		return
	}

	apRecv := p.sess.Accesspoint().Receive(ap.PacketTypeProductInfo, ap.PacketTypeCountryCode)
	msgRecv := p.sess.Dealer().ReceiveMessage("hm://pusher/v1/connections/", "hm://connect-state/v1/")
	reqRecv := p.sess.Dealer().ReceiveRequest("hm://connect-state/v1/player/command")
	playerRecv := p.audioPlayer.Receive()

	volumeTimer := time.NewTimer(time.Minute)
	volumeTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case pkt, ok := <-apRecv:
			if !ok {
				continue
			}
			if err := p.handleAccesspointPacket(pkt.Type, pkt.Payload); err != nil {
				p.log.WithError(err).Warn("failed handling accesspoint packet")
			}
		case msg, ok := <-msgRecv:
			if !ok {
				continue
			}
			if err := p.handleDealerMessage(ctx, msg); err != nil {
				p.log.WithError(err).Warn("failed handling dealer message")
			}
		case req, ok := <-reqRecv:
			if !ok {
				continue
			}
			if err := p.handleDealerRequest(ctx, req); err != nil {
				p.log.WithError(err).Warn("failed handling dealer request")
				req.Reply(false)
			} else {
				req.Reply(true)
			}
		case ev, ok := <-playerRecv:
			if !ok {
				continue
			}
			p.handlePlayerEvent(ctx, &ev)
		case <-p.prefetchTimer.C:
			p.prefetchNext(ctx)
		case volume := <-p.volumeUpdate:
			p.state.device.Volume = uint32(math.Round(float64(volume * player.MaxStateVolume)))
			volumeTimer.Reset(100 * time.Millisecond)
		case <-volumeTimer.C:
			p.volumeUpdated(ctx)
		}
	}
}

// handleAccesspointPacket processes AP packets (product info, country code).
func (p *Player) handleAccesspointPacket(pktType ap.PacketType, payload []byte) error {
	switch pktType {
	case ap.PacketTypeProductInfo:
		var prod productInfo
		if err := xml.Unmarshal(payload, &prod); err != nil {
			return fmt.Errorf("failed unmarshalling ProductInfo: %w", err)
		}
		// Store for future use (album art URLs, etc.)
		_ = prod
		return nil
	case ap.PacketTypeCountryCode:
		*p.countryCode = string(payload)
		return nil
	default:
		return nil
	}
}

type productInfo struct {
	XMLName  xml.Name `xml:"products"`
	Products []struct {
		XMLName      xml.Name `xml:"product"`
		Type         string   `xml:"type"`
		HeadFilesUrl string   `xml:"head-files-url"`
		ImageUrl     string   `xml:"image-url"`
	} `xml:"product"`
}

// handleDealerMessage processes dealer messages (connection ID, volume, cluster updates).
func (p *Player) handleDealerMessage(ctx context.Context, msg dealer.Message) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if strings.HasPrefix(msg.Uri, "hm://pusher/v1/connections/") {
		p.spotConnId = msg.Headers["Spotify-Connection-Id"]
		p.log.Debugf("received connection id")

		if err := p.putConnectState(ctx, connectpb.PutStateReason_NEW_DEVICE); err != nil {
			return fmt.Errorf("failed initial state put: %w", err)
		}
	} else if strings.HasPrefix(msg.Uri, "hm://connect-state/v1/connect/volume") {
		var setVolCmd connectpb.SetVolumeCommand
		if err := proto.Unmarshal(msg.Payload, &setVolCmd); err != nil {
			return fmt.Errorf("failed unmarshalling SetVolumeCommand: %w", err)
		}
		p.updateVolume(uint32(setVolCmd.Volume))
	} else if strings.HasPrefix(msg.Uri, "hm://connect-state/v1/connect/logout") {
		p.log.Info("received logout request")
	} else if strings.HasPrefix(msg.Uri, "hm://connect-state/v1/cluster") {
		var clusterUpdate connectpb.ClusterUpdate
		if err := proto.Unmarshal(msg.Payload, &clusterUpdate); err != nil {
			return fmt.Errorf("failed unmarshalling ClusterUpdate: %w", err)
		}

		stopBeingActive := p.state.active && clusterUpdate.Cluster.ActiveDeviceId != p.deviceId && clusterUpdate.Cluster.PlayerState.Timestamp > p.state.lastTransferTimestamp

		if !stopBeingActive {
			return nil
		}

		name := "<unknown>"
		if device := clusterUpdate.Cluster.Device[clusterUpdate.Cluster.ActiveDeviceId]; device != nil {
			name = device.Name
		}
		p.log.Infof("playback was transferred to %s", name)

		return p.stopPlayback(ctx)
	}

	return nil
}

// handleDealerRequest processes dealer requests (player commands).
func (p *Player) handleDealerRequest(ctx context.Context, req dealer.Request) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	switch req.MessageIdent {
	case "hm://connect-state/v1/player/command":
		return p.handlePlayerCommand(ctx, req.Payload)
	default:
		p.log.Warnf("unknown dealer request: %s", req.MessageIdent)
		return nil
	}
}

// handlePlayerCommand processes Spotify Connect player commands.
func (p *Player) handlePlayerCommand(ctx context.Context, req dealer.RequestPayload) error {
	p.state.lastCommand = &req

	p.log.Debugf("handling %s player command from %s", req.Command.Endpoint, req.SentByDeviceId)

	// Mark ourselves as active for any incoming command
	p.mu.Lock()
	p.active = true
	p.mu.Unlock()

	switch req.Command.Endpoint {
	case "transfer":
		if len(req.Command.Data) == 0 {
			return nil
		}

		var transferState connectpb.TransferState
		if err := proto.Unmarshal(req.Command.Data, &transferState); err != nil {
			return fmt.Errorf("failed unmarshalling TransferState: %w", err)
		}
		p.state.lastTransferTimestamp = transferState.Playback.Timestamp

		ctxTracks, err := tracks.NewTrackListFromContext(ctx, p.log, p.sess.Spclient(), transferState.CurrentSession.Context)
		if err != nil {
			return fmt.Errorf("failed creating track list: %w", err)
		}

		if sessId := transferState.CurrentSession.OriginalSessionId; sessId != nil {
			p.state.player.SessionId = *sessId
		} else {
			sessionId := make([]byte, 16)
			_, _ = rand.Read(sessionId)
			p.state.player.SessionId = base64.StdEncoding.EncodeToString(sessionId)
		}

		p.state.setActive(true)
		p.state.player.IsPlaying = false
		p.state.player.IsBuffering = false

		p.state.player.Options = transferState.Options
		pause := transferState.Playback.IsPaused && req.Command.Options.RestorePaused != "resume"
		p.state.player.Timestamp = transferState.Playback.Timestamp
		p.state.player.PositionAsOfTimestamp = int64(transferState.Playback.PositionAsOfTimestamp)
		p.state.setPaused(pause)

		p.state.player.PlayOrigin = transferState.CurrentSession.PlayOrigin
		p.state.player.PlayOrigin.DeviceIdentifier = req.SentByDeviceId
		p.state.player.ContextUri = transferState.CurrentSession.Context.Uri
		p.state.player.ContextUrl = transferState.CurrentSession.Context.Url
		p.state.player.ContextRestrictions = transferState.CurrentSession.Context.Restrictions
		p.state.player.Suppressions = transferState.CurrentSession.Suppressions

		p.state.player.ContextMetadata = map[string]string{}
		for k, v := range transferState.CurrentSession.Context.Metadata {
			p.state.player.ContextMetadata[k] = v
		}
		for k, v := range ctxTracks.Metadata() {
			p.state.player.ContextMetadata[k] = v
		}

		contextSpotType := golibrespot.InferSpotifyIdTypeFromContextUri(p.state.player.ContextUri)
		currentTrack := golibrespot.ContextTrackToProvidedTrack(contextSpotType, transferState.Playback.CurrentTrack)
		if err := ctxTracks.TrySeek(ctx, tracks.ProvidedTrackComparator(contextSpotType, currentTrack)); err != nil {
			return fmt.Errorf("failed seeking to track: %w", err)
		}

		if err := ctxTracks.ToggleShuffle(ctx, transferState.Options.ShufflingContext); err != nil {
			return fmt.Errorf("failed shuffling context")
		}

		p.state.queueID = 0
		for _, track := range transferState.Queue.Tracks {
			if track.Uid == "" || track.Uid[0] != 'q' {
				continue
			}
			n, err := strconv.ParseUint(track.Uid[1:], 10, 64)
			if err != nil {
				continue
			}
			p.state.queueID = max(p.state.queueID, n)
		}

		for _, track := range transferState.Queue.Tracks {
			ctxTracks.AddToQueue(track)
		}
		ctxTracks.SetPlayingQueue(transferState.Queue.IsPlayingQueue)

		p.state.tracks = ctxTracks
		p.state.player.Track = ctxTracks.CurrentTrack()
		p.state.player.PrevTracks = ctxTracks.PrevTracks()
		p.state.player.NextTracks = ctxTracks.NextTracks(ctx, nil)
		p.state.player.Index = ctxTracks.Index()

		if err := p.loadCurrentTrack(ctx, pause, true); err != nil {
			return fmt.Errorf("failed loading current track (transfer): %w", err)
		}

		return nil

	case "play":
		p.state.setActive(true)
		p.state.player.PlayOrigin = req.Command.PlayOrigin
		p.state.player.PlayOrigin.DeviceIdentifier = req.SentByDeviceId
		p.state.player.Suppressions = req.Command.Options.Suppressions

		if req.Command.Options.PlayerOptionsOverride != nil {
			p.state.player.Options.ShufflingContext = req.Command.Options.PlayerOptionsOverride.ShufflingContext
			p.state.player.Options.RepeatingTrack = req.Command.Options.PlayerOptionsOverride.RepeatingTrack
			p.state.player.Options.RepeatingContext = req.Command.Options.PlayerOptionsOverride.RepeatingContext
		}

		var skipTo skipToFunc
		if len(req.Command.Options.SkipTo.TrackUri) > 0 || len(req.Command.Options.SkipTo.TrackUid) > 0 || req.Command.Options.SkipTo.TrackIndex > 0 {
			index := -1
			skipTo = func(track *connectpb.ContextTrack) bool {
				if len(req.Command.Options.SkipTo.TrackUid) > 0 && req.Command.Options.SkipTo.TrackUid == track.Uid {
					return true
				} else if len(req.Command.Options.SkipTo.TrackUri) > 0 && req.Command.Options.SkipTo.TrackUri == track.Uri {
					return true
				} else if req.Command.Options.SkipTo.TrackIndex != 0 && len(req.Command.Options.SkipTo.TrackUri) == 0 && len(req.Command.Options.SkipTo.TrackUid) == 0 {
					index += 1
					return index == req.Command.Options.SkipTo.TrackIndex
				}
				return false
			}
		}

		return p.loadContext(ctx, req.Command.Context, skipTo, req.Command.Options.InitiallyPaused, true)

	case "pause":
		return p.pause(ctx)
	case "resume":
		return p.play(ctx)
	case "seek_to":
		var position int64
		if req.Command.Relative == "current" {
			position = p.audioPlayer.PositionMs() + req.Command.Position
		} else if req.Command.Relative == "beginning" || req.Command.Relative == "" {
			if req.Command.Relative == "" {
				if pos, ok := req.Command.Value.(float64); ok {
					position = int64(pos)
				} else {
					return nil
				}
			} else {
				position = req.Command.Position
			}
		} else {
			return nil
		}
		return p.seek(ctx, position)
	case "skip_prev":
		return p.skipPrev(ctx, req.Command.Options.AllowSeeking)
	case "skip_next":
		return p.skipNext(ctx, req.Command.Track)
	case "update_context":
		if req.Command.Context.Uri != p.state.player.ContextUri {
			return nil
		}
		p.state.player.ContextRestrictions = req.Command.Context.Restrictions
		if p.state.player.ContextMetadata == nil {
			p.state.player.ContextMetadata = map[string]string{}
		}
		for k, v := range req.Command.Context.Metadata {
			p.state.player.ContextMetadata[k] = v
		}
		p.updateState(ctx)
		return nil
	case "set_repeating_context":
		val := req.Command.Value.(bool)
		p.setOptions(ctx, &val, nil, nil)
		return nil
	case "set_repeating_track":
		val := req.Command.Value.(bool)
		p.setOptions(ctx, nil, &val, nil)
		return nil
	case "set_shuffling_context":
		val := req.Command.Value.(bool)
		p.setOptions(ctx, nil, nil, &val)
		return nil
	case "set_options":
		p.setOptions(ctx, req.Command.RepeatingContext, req.Command.RepeatingTrack, req.Command.ShufflingContext)
		return nil
	case "set_queue":
		p.setQueue(ctx, req.Command.PrevTracks, req.Command.NextTracks)
		return nil
	case "add_to_queue":
		p.addToQueue(ctx, req.Command.Track)
		return nil
	default:
		return fmt.Errorf("unsupported player command: %s", req.Command.Endpoint)
	}
}

// emitEvent broadcasts a player event to all subscribers.
func (p *Player) emitEvent(eventType string) {
	state := p.GetState()
	event := &gspotv1.SubscribeResponse{
		State:     state,
		EventType: eventType,
	}

	p.subscribersMu.RLock()
	defer p.subscribersMu.RUnlock()

	for _, ch := range p.subscribers {
		// Bounded channel — drop oldest if full.
		select {
		case ch <- event:
		default:
			// Channel full, drop oldest and try again.
			select {
			case <-ch:
			default:
			}
			select {
			case ch <- event:
			default:
			}
		}
	}
}

// startZeroconf starts the Spotify Connect zeroconf discovery.
func (p *Player) startZeroconf(ctx context.Context) {
	z, err := zeroconf.NewZeroconf(p.log, 0, p.conf.Librespot.DeviceName, p.deviceId, p.deviceType, nil, false)
	if err != nil {
		p.log.WithError(err).Error("failed initializing zeroconf")
		return
	}

	z.SetCurrentUser(p.sess.Username())

	_ = z.Serve(func(req zeroconf.NewUserRequest) bool {
		p.log.WithField("username", golibrespot.ObfuscateUsername(req.Username)).
			Infof("zeroconf connection from %s", req.DeviceName)
		// For now, reject new zeroconf users — we only support stored credentials.
		// TODO: support zeroconf user switching
		return false
	})
}

// repeatStateString converts proto options to a string.
func repeatStateString(opts *connectpb.ContextPlayerOptions) string {
	if opts == nil {
		return "off"
	}
	if opts.RepeatingTrack {
		return "track"
	}
	if opts.RepeatingContext {
		return "context"
	}
	return "off"
}

// Unused import suppression
var _ = bytes.Compare
