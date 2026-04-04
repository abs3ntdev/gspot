package librespot

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	golibrespot "github.com/devgianlu/go-librespot"
	"github.com/devgianlu/go-librespot/player"
	connectpb "github.com/devgianlu/go-librespot/proto/spotify/connectstate"
	playerpb "github.com/devgianlu/go-librespot/proto/spotify/player"
	"github.com/devgianlu/go-librespot/tracks"
	"google.golang.org/protobuf/proto"
)

func (p *Player) prefetchNext(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	next := p.state.tracks.PeekNext(ctx)
	if next == nil {
		return
	}

	if next.Uri == "" {
		p.log.Warn("cannot prefetch next track because the uri field is empty")
		return
	}

	nextId, err := golibrespot.SpotifyIdFromUri(next.Uri)
	if err != nil {
		p.log.WithError(err).WithField("uri", next.Uri).Warn("failed parsing prefetch uri")
		return
	} else if p.secondaryStream != nil && p.secondaryStream.Is(*nextId) {
		return
	}

	p.log.WithField("uri", nextId.Uri()).Debugf("prefetching next %s", nextId.Type())

	p.secondaryStream, err = p.audioPlayer.NewStream(ctx, p.httpClient, *nextId, p.conf.Librespot.Bitrate, 0)
	if err != nil {
		p.log.WithError(err).WithField("uri", nextId.String()).Warnf("failed prefetching %s stream", nextId.Type())
		return
	}

	p.audioPlayer.SetSecondaryStream(p.secondaryStream.Source)

	p.log.WithField("uri", nextId.Uri()).
		Infof("prefetched %s %s (duration: %dms)", nextId.Type(),
			strconv.QuoteToGraphic(p.secondaryStream.Media.Name()), p.secondaryStream.Media.Duration())
}

func (p *Player) schedulePrefetchNext() {
	if p.state.player.IsPaused || p.primaryStream == nil {
		p.prefetchTimer.Stop()
		return
	}

	untilTrackEnd := time.Duration(p.primaryStream.Media.Duration()-int32(p.audioPlayer.PositionMs())) * time.Millisecond
	untilTrackEnd -= 30 * time.Second
	if untilTrackEnd < 10*time.Second {
		p.prefetchTimer.Reset(0)
	} else {
		p.prefetchTimer.Reset(untilTrackEnd)
	}
}

func (p *Player) handlePlayerEvent(ctx context.Context, ev *player.Event) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	switch ev.Type {
	case player.EventTypePlay:
		p.state.player.IsPlaying = true
		p.state.setPaused(false)
		p.state.player.IsBuffering = false
		p.updateState(ctx)

		p.sess.Events().OnPlayerPlay(
			p.primaryStream,
			p.state.player.ContextUri,
			p.state.player.Options.ShufflingContext,
			p.state.player.PlayOrigin,
			p.state.tracks.CurrentTrack(),
			p.state.trackPosition(),
		)

		p.emitEvent("playing")

	case player.EventTypeResume:
		p.state.player.IsPlaying = true
		p.state.setPaused(false)
		p.state.player.IsBuffering = false
		p.updateState(ctx)

		p.sess.Events().OnPlayerResume(p.primaryStream, p.state.trackPosition())

		p.emitEvent("playing")

	case player.EventTypePause:
		p.state.player.IsPlaying = true
		p.state.setPaused(true)
		p.state.player.IsBuffering = false
		p.updateState(ctx)

		p.sess.Events().OnPlayerPause(
			p.primaryStream,
			p.state.player.ContextUri,
			p.state.player.Options.ShufflingContext,
			p.state.player.PlayOrigin,
			p.state.tracks.CurrentTrack(),
			p.state.trackPosition(),
		)

		p.emitEvent("paused")

	case player.EventTypeNotPlaying:
		p.sess.Events().OnPlayerEnd(p.primaryStream, p.state.trackPosition())

		hasNextTrack, err := p.advanceNext(context.TODO(), false, false)
		if err != nil {
			p.log.WithError(err).Error("failed advancing to next track")
		}

		if !hasNextTrack {
			p.emitEvent("stopped")
		}

	case player.EventTypeStop:
		p.emitEvent("stopped")

	default:
		p.log.Warnf("unhandled player event type: %d", ev.Type)
	}
}

type skipToFunc func(*connectpb.ContextTrack) bool

func (p *Player) loadContext(ctx context.Context, spotCtx *connectpb.Context, skipTo skipToFunc, paused, drop bool) error {
	ctxTracks, err := tracks.NewTrackListFromContext(ctx, p.log, p.sess.Spclient(), spotCtx)
	if err != nil {
		return fmt.Errorf("failed creating track list: %w", err)
	}

	p.state.setPaused(paused)

	sessionId := make([]byte, 16)
	_, _ = rand.Read(sessionId)
	p.state.player.SessionId = base64.StdEncoding.EncodeToString(sessionId)

	p.state.player.ContextUri = spotCtx.Uri
	p.state.player.ContextUrl = spotCtx.Url
	p.state.player.Restrictions = spotCtx.Restrictions
	p.state.player.ContextRestrictions = spotCtx.Restrictions

	if spotCtx.Restrictions != nil {
		if len(spotCtx.Restrictions.DisallowTogglingShuffleReasons) > 0 {
			p.state.player.Options.ShufflingContext = false
		}
		if len(spotCtx.Restrictions.DisallowTogglingRepeatTrackReasons) > 0 {
			p.state.player.Options.RepeatingTrack = false
		}
		if len(spotCtx.Restrictions.DisallowTogglingRepeatContextReasons) > 0 {
			p.state.player.Options.RepeatingContext = false
		}
	}

	if p.state.player.ContextMetadata == nil {
		p.state.player.ContextMetadata = map[string]string{}
	}
	for k, v := range spotCtx.Metadata {
		p.state.player.ContextMetadata[k] = v
	}

	p.state.player.Timestamp = time.Now().UnixMilli()
	p.state.player.PositionAsOfTimestamp = 0

	if skipTo == nil {
		if err := ctxTracks.ToggleShuffle(ctx, p.state.player.Options.ShufflingContext); err != nil {
			return fmt.Errorf("failed shuffling context")
		}
		if err := ctxTracks.TrySeek(ctx, func(_ *connectpb.ContextTrack) bool { return true }); err != nil {
			return fmt.Errorf("failed seeking to track: %w", err)
		}
	} else {
		if err := ctxTracks.TrySeek(ctx, skipTo); err != nil {
			return fmt.Errorf("failed seeking to track: %w", err)
		}
		if err := ctxTracks.ToggleShuffle(ctx, p.state.player.Options.ShufflingContext); err != nil {
			return fmt.Errorf("failed shuffling context")
		}
	}

	p.state.tracks = ctxTracks
	p.state.player.Track = ctxTracks.CurrentTrack()
	p.state.player.PrevTracks = ctxTracks.PrevTracks()
	p.state.player.NextTracks = ctxTracks.NextTracks(ctx, nil)
	p.state.player.Index = ctxTracks.Index()

	if err := p.loadCurrentTrack(ctx, paused, drop); err != nil {
		return fmt.Errorf("failed loading current track (load context): %w", err)
	}

	return nil
}

func (p *Player) loadCurrentTrack(ctx context.Context, paused, drop bool) error {
	if p.primaryStream != nil {
		p.sess.Events().OnPrimaryStreamUnload(p.primaryStream, p.audioPlayer.PositionMs())
		p.primaryStream = nil
	}

	spotId, err := golibrespot.SpotifyIdFromUri(p.state.player.Track.Uri)
	if err != nil {
		return fmt.Errorf("failed parsing uri: %w", err)
	} else if spotId.Type() != golibrespot.SpotifyIdTypeTrack && spotId.Type() != golibrespot.SpotifyIdTypeEpisode {
		return fmt.Errorf("unsupported spotify type: %s", spotId.Type())
	}

	trackPosition := p.state.trackPosition()
	p.log.WithField("uri", spotId.Uri()).
		Debugf("loading %s (paused: %t, position: %dms)", spotId.Type(), paused, trackPosition)

	p.state.updateTimestamp()
	p.state.player.IsPlaying = true
	p.state.player.IsBuffering = true
	p.state.player.IsPaused = paused
	p.state.player.PlaybackSpeed = 0
	p.updateState(ctx)

	if p.secondaryStream != nil && p.secondaryStream.Is(*spotId) {
		p.primaryStream = p.secondaryStream
		p.secondaryStream = nil
	} else {
		p.secondaryStream = nil
		p.primaryStream, err = p.audioPlayer.NewStream(ctx, p.httpClient, *spotId, p.conf.Librespot.Bitrate, trackPosition)
		if err != nil {
			return fmt.Errorf("failed creating stream for %s: %w", spotId, err)
		}
	}

	if err := p.audioPlayer.SetPrimaryStream(p.primaryStream.Source, paused, drop); err != nil {
		return fmt.Errorf("failed setting stream for %s: %w", spotId, err)
	}

	p.sess.Events().PostPrimaryStreamLoad(p.primaryStream, paused)

	p.log.WithField("uri", spotId.Uri()).
		Infof("loaded %s %s (paused: %t, position: %dms, duration: %dms)", spotId.Type(),
			strconv.QuoteToGraphic(p.primaryStream.Media.Name()), paused, trackPosition, p.primaryStream.Media.Duration())

	p.state.updateTimestamp()
	p.state.player.PlaybackId = hex.EncodeToString(p.primaryStream.PlaybackId)
	p.state.player.Duration = int64(p.primaryStream.Media.Duration())
	p.state.player.IsPlaying = true
	p.state.player.IsBuffering = false
	p.state.setPaused(paused)
	p.updateState(ctx)
	p.schedulePrefetchNext()

	p.emitEvent("track_changed")
	return nil
}

func (p *Player) setOptions(ctx context.Context, repeatingContext *bool, repeatingTrack *bool, shufflingContext *bool) {
	var requiresUpdate bool
	if repeatingContext != nil && *repeatingContext != p.state.player.Options.RepeatingContext {
		p.state.player.Options.RepeatingContext = *repeatingContext
		requiresUpdate = true
	}

	if repeatingTrack != nil && *repeatingTrack != p.state.player.Options.RepeatingTrack {
		p.state.player.Options.RepeatingTrack = *repeatingTrack
		requiresUpdate = true
	}

	if p.state.tracks != nil && shufflingContext != nil && *shufflingContext != p.state.player.Options.ShufflingContext {
		if err := p.state.tracks.ToggleShuffle(ctx, *shufflingContext); err != nil {
			p.log.WithError(err).Errorf("failed toggling shuffle context (value: %t)", *shufflingContext)
			return
		}

		p.state.player.Options.ShufflingContext = *shufflingContext
		p.state.player.Track = p.state.tracks.CurrentTrack()
		p.state.player.PrevTracks = p.state.tracks.PrevTracks()
		p.state.player.NextTracks = p.state.tracks.NextTracks(ctx, nil)
		p.state.player.Index = p.state.tracks.Index()
		requiresUpdate = true
	}

	if requiresUpdate {
		p.updateState(ctx)
	}
}

func (p *Player) addToQueue(ctx context.Context, track *connectpb.ContextTrack) {
	if p.state.tracks == nil {
		p.log.Warnf("cannot add to queue without a context")
		return
	}

	if track.Uid == "" {
		p.state.queueID++
		track.Uid = fmt.Sprintf("q%d", p.state.queueID)
	}

	p.state.tracks.AddToQueue(track)
	p.state.player.PrevTracks = p.state.tracks.PrevTracks()
	p.state.player.NextTracks = p.state.tracks.NextTracks(ctx, nil)
	p.updateState(ctx)
	p.schedulePrefetchNext()
}

func (p *Player) setQueue(ctx context.Context, prev []*connectpb.ContextTrack, next []*connectpb.ContextTrack) {
	if p.state.tracks == nil {
		p.log.Warnf("cannot set queue without a context")
		return
	}

	p.state.tracks.SetQueue(prev, next)
	p.state.player.PrevTracks = p.state.tracks.PrevTracks()
	p.state.player.NextTracks = p.state.tracks.NextTracks(ctx, next)
	p.updateState(ctx)
	p.schedulePrefetchNext()
}

func (p *Player) play(ctx context.Context) error {
	if p.primaryStream == nil {
		return fmt.Errorf("no primary stream")
	}

	seekPos := p.state.trackPosition()
	seekPos = max(0, min(seekPos, int64(p.primaryStream.Media.Duration())))
	if err := p.audioPlayer.SeekMs(seekPos); err != nil {
		return fmt.Errorf("failed seeking before play: %w", err)
	}

	if err := p.audioPlayer.Play(); err != nil {
		return fmt.Errorf("failed starting playback: %w", err)
	}

	streamPos := p.audioPlayer.PositionMs()
	p.state.player.Timestamp = time.Now().UnixMilli()
	p.state.player.PositionAsOfTimestamp = streamPos
	p.state.setPaused(false)
	p.updateState(ctx)
	p.schedulePrefetchNext()

	return nil
}

func (p *Player) pause(ctx context.Context) error {
	if p.primaryStream == nil {
		return fmt.Errorf("no primary stream")
	}

	streamPos := p.audioPlayer.PositionMs()

	if err := p.audioPlayer.Pause(); err != nil {
		return fmt.Errorf("failed pausing playback: %w", err)
	}

	p.state.player.Timestamp = time.Now().UnixMilli()
	p.state.player.PositionAsOfTimestamp = streamPos
	p.state.setPaused(true)
	p.updateState(ctx)
	p.schedulePrefetchNext()

	return nil
}

func (p *Player) seek(ctx context.Context, position int64) error {
	if p.primaryStream == nil {
		return fmt.Errorf("no primary stream")
	}

	oldPosition := p.audioPlayer.PositionMs()
	position = max(0, min(position, int64(p.primaryStream.Media.Duration())))

	if err := p.audioPlayer.SeekMs(position); err != nil {
		return err
	}

	p.state.player.Timestamp = time.Now().UnixMilli()
	p.state.player.PositionAsOfTimestamp = position
	p.updateState(ctx)
	p.schedulePrefetchNext()

	p.sess.Events().OnPlayerSeek(p.primaryStream, oldPosition, position)

	p.emitEvent("seek")
	return nil
}

func (p *Player) skipPrev(ctx context.Context, allowSeeking bool) error {
	if allowSeeking && p.audioPlayer.PositionMs() > 3000 {
		return p.seek(ctx, 0)
	}

	p.sess.Events().OnPlayerSkipBackward(p.primaryStream, p.audioPlayer.PositionMs())

	if p.state.tracks != nil {
		p.state.tracks.GoPrev()

		p.state.player.Track = p.state.tracks.CurrentTrack()
		p.state.player.PrevTracks = p.state.tracks.PrevTracks()
		p.state.player.NextTracks = p.state.tracks.NextTracks(ctx, nil)
		p.state.player.Index = p.state.tracks.Index()
	}

	p.state.player.Timestamp = time.Now().UnixMilli()
	p.state.player.PositionAsOfTimestamp = 0

	if err := p.loadCurrentTrack(ctx, p.state.player.IsPaused, true); err != nil {
		return fmt.Errorf("failed loading current track (skip prev): %w", err)
	}

	return nil
}

func (p *Player) skipNext(ctx context.Context, track *connectpb.ContextTrack) error {
	p.sess.Events().OnPlayerSkipForward(p.primaryStream, p.audioPlayer.PositionMs(), track != nil)

	if track != nil {
		contextSpotType := golibrespot.InferSpotifyIdTypeFromContextUri(p.state.player.ContextUri)
		if err := p.state.tracks.TrySeek(ctx, tracks.ContextTrackComparator(contextSpotType, track)); err != nil {
			return err
		}

		p.state.player.Timestamp = time.Now().UnixMilli()
		p.state.player.PositionAsOfTimestamp = 0

		p.state.player.Track = p.state.tracks.CurrentTrack()
		p.state.player.PrevTracks = p.state.tracks.PrevTracks()
		p.state.player.NextTracks = p.state.tracks.NextTracks(ctx, nil)
		p.state.player.Index = p.state.tracks.Index()

		if err := p.loadCurrentTrack(ctx, p.state.player.IsPaused, true); err != nil {
			return err
		}
		return nil
	}

	hasNextTrack, err := p.advanceNext(ctx, true, true)
	if err != nil {
		return fmt.Errorf("failed skipping to next track: %w", err)
	}

	if !hasNextTrack {
		p.emitEvent("stopped")
	}

	return nil
}

func (p *Player) advanceNext(ctx context.Context, forceNext, drop bool) (bool, error) {
	var uri string
	var hasNextTrack bool
	if p.state.tracks != nil {
		if !forceNext && p.state.player.Options.RepeatingTrack {
			hasNextTrack = true
			p.state.player.IsPaused = false
		} else {
			hasNextTrack = p.state.tracks.GoNext(ctx)

			if !hasNextTrack {
				hasNextTrack = p.state.tracks.GoStart(ctx)
				if !p.state.player.Options.RepeatingContext {
					hasNextTrack = false
				}
			}

			p.state.player.IsPaused = !hasNextTrack
		}

		p.state.player.Track = p.state.tracks.CurrentTrack()
		p.state.player.PrevTracks = p.state.tracks.PrevTracks()
		p.state.player.NextTracks = p.state.tracks.NextTracks(ctx, nil)
		p.state.player.Index = p.state.tracks.Index()

		uri = p.state.player.Track.Uri
	}

	p.state.player.Timestamp = time.Now().UnixMilli()
	p.state.player.PositionAsOfTimestamp = 0

	if !hasNextTrack && p.conf.Librespot.Autoplay && !strings.HasPrefix(p.state.player.ContextUri, "spotify:station:") {
		p.state.player.Suppressions = &connectpb.Suppressions{}

		var prevTrackUris []string
		if p.state.tracks != nil {
			for _, track := range p.state.tracks.AllTracks(ctx) {
				prevTrackUris = append(prevTrackUris, track.Uri)
			}
		}

		if len(prevTrackUris) == 0 {
			p.log.Warnf("cannot resolve autoplay station because there are no previous tracks")
			return false, nil
		}

		spotCtx, err := p.sess.Spclient().ContextResolveAutoplay(ctx, &playerpb.AutoplayContextRequest{
			ContextUri:     proto.String(p.state.player.ContextUri),
			RecentTrackUri: prevTrackUris,
		})
		if err != nil {
			p.log.WithError(err).Warnf("failed resolving autoplay station")
			return false, nil
		}

		if err := p.loadContext(ctx, spotCtx, func(_ *connectpb.ContextTrack) bool { return true }, false, drop); err != nil {
			p.log.WithError(err).Warnf("failed loading autoplay station")
			return false, nil
		}

		return true, nil
	}

	if !hasNextTrack {
		p.state.player.IsPlaying = false
		p.state.player.IsPaused = false
		p.state.player.IsBuffering = false
	}

	if err := p.loadCurrentTrack(ctx, !hasNextTrack, drop); errors.Is(err, golibrespot.ErrMediaRestricted) || errors.Is(err, golibrespot.ErrNoSupportedFormats) {
		p.log.WithError(err).Infof("skipping unplayable media: %s", uri)
		if forceNext {
			return false, err
		}
		return p.advanceNext(ctx, true, drop)
	} else if err != nil {
		return false, fmt.Errorf("failed loading current track (advance to %s): %w", uri, err)
	}

	return hasNextTrack, nil
}

func (p *Player) updateVolume(newVal uint32) {
	if newVal > player.MaxStateVolume {
		newVal = player.MaxStateVolume
	}

	p.audioPlayer.SetVolume(newVal)

	p.appState.LastVolume = &newVal
	if err := p.appState.Write(); err != nil {
		p.log.WithError(err).Error("failed writing state after volume change")
	}

	// Drain the channel if there's a pending value.
	select {
	case <-p.volumeUpdate:
	default:
	}

	p.volumeUpdate <- float32(newVal) / player.MaxStateVolume
}

func (p *Player) volumeUpdated(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := p.putConnectState(ctx, connectpb.PutStateReason_VOLUME_CHANGED); err != nil {
		p.log.WithError(err).Error("failed put state after volume change")
	}

	p.emitEvent("volume_changed")
}

func (p *Player) stopPlayback(ctx context.Context) error {
	p.audioPlayer.Stop()
	p.primaryStream = nil
	p.secondaryStream = nil

	p.state.reset()
	if err := p.putConnectState(ctx, connectpb.PutStateReason_BECAME_INACTIVE); err != nil {
		return fmt.Errorf("failed inactive state put: %w", err)
	}

	p.schedulePrefetchNext()

	p.mu.Lock()
	p.active = false
	p.mu.Unlock()

	p.emitEvent("stopped")
	return nil
}

func (p *Player) apiVolume() uint32 {
	return uint32(math.Round(float64(p.state.device.Volume*p.conf.Librespot.VolumeSteps) / player.MaxStateVolume))
}
