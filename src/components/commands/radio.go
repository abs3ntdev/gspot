package commands

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/zmb3/spotify/v2"
	_ "modernc.org/sqlite"
)

func (c *Commander) Radio() error {
	currentSong, err := c.Client().PlayerCurrentlyPlaying(c.Context)
	if err != nil {
		return err
	}
	if currentSong.Item != nil {
		return c.RadioGivenSong(currentSong.Item.SimpleTrack, currentSong.Progress)
	}
	_, err = c.activateDevice()
	if err != nil {
		return err
	}
	tracks, err := c.Client().CurrentUsersTracks(c.Context, spotify.Limit(10))
	if err != nil {
		return err
	}
	return c.RadioGivenSong(tracks.Tracks[rand.Intn(len(tracks.Tracks))].SimpleTrack, 0)
}

func (c *Commander) RadioFromPlaylist(playlist spotify.SimplePlaylist) error {
	playlistPage, err := c.Client().GetPlaylistItems(
		c.Context,
		playlist.ID,
		spotify.Limit(50),
		spotify.Offset(0),
	)
	if err != nil {
		return err
	}
	pageSongs := playlistPage.Items
	rand.Shuffle(len(pageSongs), func(i, j int) { pageSongs[i], pageSongs[j] = pageSongs[j], pageSongs[i] })
	seedCount := 5
	if len(pageSongs) < seedCount {
		seedCount = len(pageSongs)
	}
	seedIds := []spotify.ID{}
	for idx, song := range pageSongs {
		if idx >= seedCount {
			break
		}
		seedIds = append(seedIds, song.Track.Track.ID)
	}
	return c.RadioGivenList(seedIds[:seedCount], playlist.Name)
}

func (c *Commander) RadioFromSavedTracks() error {
	savedSongs, err := c.Client().CurrentUsersTracks(c.Context, spotify.Limit(50), spotify.Offset(0))
	if err != nil {
		return err
	}
	if savedSongs.Total == 0 {
		return fmt.Errorf("you have no saved songs")
	}
	pages := int(math.Ceil(float64(savedSongs.Total) / 50))
	randomPage := 1
	if pages > 1 {
		randomPage = rand.Intn(pages-1) + 1
	}
	trackPage, err := c.Client().CurrentUsersTracks(c.Context, spotify.Limit(50), spotify.Offset(randomPage*50))
	if err != nil {
		return err
	}
	pageSongs := trackPage.Tracks
	rand.Shuffle(len(pageSongs), func(i, j int) { pageSongs[i], pageSongs[j] = pageSongs[j], pageSongs[i] })
	seedCount := 4
	seedIds := []spotify.ID{}
	for idx, song := range pageSongs {
		if idx >= seedCount {
			break
		}
		seedIds = append(seedIds, song.ID)
	}
	seedIds = append(seedIds, savedSongs.Tracks[0].ID)
	return c.RadioGivenList(seedIds, "Saved Tracks")
}

func (c *Commander) RadioGivenArtist(artist spotify.SimpleArtist) error {
	seed := spotify.Seeds{
		Artists: []spotify.ID{artist.ID},
	}
	recomendations, err := c.Client().GetRecommendations(c.Context, seed, &spotify.TrackAttributes{}, spotify.Limit(100))
	if err != nil {
		return err
	}
	recomendationIds := []spotify.ID{}
	for _, song := range recomendations.Tracks {
		recomendationIds = append(recomendationIds, song.ID)
	}
	err = c.ClearRadio()
	if err != nil {
		return err
	}
	radioPlaylist, db, err := c.GetRadioPlaylist(artist.Name)
	if err != nil {
		return err
	}
	queue := []spotify.ID{}
	for _, rec := range recomendationIds {
		exists, err := c.SongExists(db, rec)
		if err != nil {
			return err
		}
		if !exists {
			_, err := db.QueryContext(c.Context, fmt.Sprintf("INSERT INTO radio (id) VALUES('%s')", string(rec)))
			if err != nil {
				return err
			}
			queue = append(queue, rec)
		}
	}
	_, err = c.Client().AddTracksToPlaylist(c.Context, radioPlaylist.ID, queue...)
	if err != nil {
		return err
	}
	err = c.Client().PlayOpt(c.Context, &spotify.PlayOptions{
		PlaybackContext: &radioPlaylist.URI,
	})
	if err != nil {
		return err
	}
	err = c.Client().Repeat(c.Context, "context")
	if err != nil {
		return err
	}
	for i := 0; i < 4; i++ {
		id := rand.Intn(len(recomendationIds)-2) + 1
		seed := spotify.Seeds{
			Tracks: []spotify.ID{recomendationIds[id]},
		}
		additionalRecs, err := c.Client().
			GetRecommendations(c.Context, seed, &spotify.TrackAttributes{}, spotify.Limit(100))
		if err != nil {
			return err
		}
		additionalRecsIds := []spotify.ID{}
		for _, song := range additionalRecs.Tracks {
			exists, err := c.SongExists(db, song.ID)
			if err != nil {
				return err
			}
			if !exists {
				_, err = db.QueryContext(c.Context, fmt.Sprintf("INSERT INTO radio (id) VALUES('%s')", string(song.ID)))
				if err != nil {
					return err
				}
				additionalRecsIds = append(additionalRecsIds, song.ID)
			}
		}
		_, err = c.Client().AddTracksToPlaylist(c.Context, radioPlaylist.ID, additionalRecsIds...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Commander) RadioGivenSong(song spotify.SimpleTrack, pos spotify.Numeric) error {
	start := time.Now().UnixMilli()
	seed := spotify.Seeds{
		Tracks: []spotify.ID{song.ID},
	}
	recomendations, err := c.Client().GetRecommendations(c.Context, seed, &spotify.TrackAttributes{}, spotify.Limit(99))
	if err != nil {
		return err
	}
	recomendationIds := []spotify.ID{}
	for _, song := range recomendations.Tracks {
		recomendationIds = append(recomendationIds, song.ID)
	}
	err = c.ClearRadio()
	if err != nil {
		return err
	}
	radioPlaylist, db, err := c.GetRadioPlaylist(song.Name)
	if err != nil {
		return err
	}
	_, err = db.QueryContext(c.Context, fmt.Sprintf("INSERT INTO radio (id) VALUES('%s')", string(song.ID)))
	if err != nil {
		return err
	}
	queue := []spotify.ID{song.ID}
	for _, rec := range recomendationIds {
		exists, err := c.SongExists(db, rec)
		if err != nil {
			return err
		}
		if !exists {
			_, err := db.QueryContext(c.Context, fmt.Sprintf("INSERT INTO radio (id) VALUES('%s')", string(rec)))
			if err != nil {
				return err
			}
			queue = append(queue, rec)
		}
	}
	_, err = c.Client().AddTracksToPlaylist(c.Context, radioPlaylist.ID, queue...)
	if err != nil {
		return err
	}
	delay := time.Now().UnixMilli() - start
	if pos != 0 {
		pos = pos + spotify.Numeric(delay)
	}
	err = c.Client().PlayOpt(c.Context, &spotify.PlayOptions{
		PlaybackContext: &radioPlaylist.URI,
		PositionMs:      pos,
	})
	if err != nil {
		if isNoActiveError(err) {
			deviceID, err := c.activateDevice()
			if err != nil {
				return err
			}
			err = c.Client().PlayOpt(c.Context, &spotify.PlayOptions{
				PlaybackContext: &radioPlaylist.URI,
				DeviceID:        &deviceID,
				PositionMs:      pos,
			})
			if err != nil {
				return err
			}
		}
	}
	err = c.Client().Repeat(c.Context, "context")
	if err != nil {
		return err
	}
	for i := 0; i < 4; i++ {
		id := rand.Intn(len(recomendationIds)-2) + 1
		seed := spotify.Seeds{
			Tracks: []spotify.ID{recomendationIds[id]},
		}
		additionalRecs, err := c.Client().
			GetRecommendations(c.Context, seed, &spotify.TrackAttributes{}, spotify.Limit(100))
		if err != nil {
			return err
		}
		additionalRecsIds := []spotify.ID{}
		for _, song := range additionalRecs.Tracks {
			exists, err := c.SongExists(db, song.ID)
			if err != nil {
				return err
			}
			if !exists {
				_, err = db.QueryContext(c.Context, fmt.Sprintf("INSERT INTO radio (id) VALUES('%s')", string(song.ID)))
				if err != nil {
					return err
				}
				additionalRecsIds = append(additionalRecsIds, song.ID)
			}
		}
		_, err = c.Client().AddTracksToPlaylist(c.Context, radioPlaylist.ID, additionalRecsIds...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Commander) ClearRadio() error {
	radioPlaylist, db, err := c.GetRadioPlaylist("")
	if err != nil {
		return err
	}
	err = c.Client().UnfollowPlaylist(c.Context, radioPlaylist.ID)
	if err != nil {
		return err
	}
	_, _ = db.Query("DROP TABLE IF EXISTS radio")
	configDir, _ := os.UserConfigDir()
	os.Remove(filepath.Join(configDir, "gspot/radio.json"))
	_ = c.Client().Pause(c.Context)
	return nil
}

func (c *Commander) GetRadioPlaylist(name string) (*spotify.FullPlaylist, *sql.DB, error) {
	configDir, _ := os.UserConfigDir()
	playlistFile, err := os.ReadFile(filepath.Join(configDir, "gspot/radio.json"))
	if errors.Is(err, os.ErrNotExist) {
		return c.CreateRadioPlaylist(name)
	}
	if err != nil {
		return nil, nil, err
	}
	var playlist *spotify.FullPlaylist
	err = json.Unmarshal(playlistFile, &playlist)
	if err != nil {
		return nil, nil, err
	}
	db, err := sql.Open("sqlite", filepath.Join(configDir, "gspot/radio.db"))
	if err != nil {
		return nil, nil, err
	}
	return playlist, db, nil
}

func (c *Commander) CreateRadioPlaylist(name string) (*spotify.FullPlaylist, *sql.DB, error) {
	// private flag doesnt work
	configDir, _ := os.UserConfigDir()
	playlist, err := c.Client().
		CreatePlaylistForUser(c.Context, c.User.ID, name+" - autoradio", "Automanaged radio playlist", false, false)
	if err != nil {
		return nil, nil, err
	}
	raw, err := json.MarshalIndent(playlist, "", " ")
	if err != nil {
		return nil, nil, err
	}
	err = os.WriteFile(filepath.Join(configDir, "gspot/radio.json"), raw, 0o600)
	if err != nil {
		return nil, nil, err
	}
	db, err := sql.Open("sqlite", filepath.Join(configDir, "gspot/radio.db"))
	if err != nil {
		return nil, nil, err
	}
	_, _ = db.QueryContext(c.Context, "DROP TABLE IF EXISTS radio")
	_, _ = db.QueryContext(c.Context, "CREATE TABLE IF NOT EXISTS radio (id string PRIMARY KEY)")
	return playlist, db, nil
}

func (c *Commander) SongExists(db *sql.DB, song spotify.ID) (bool, error) {
	songID := string(song)
	sqlStmt := `SELECT id FROM radio WHERE id = ?`
	err := db.QueryRow(sqlStmt, songID).Scan(&songID)
	if err != nil {
		if err != sql.ErrNoRows {
			return false, err
		}

		return false, nil
	}

	return true, nil
}

func (c *Commander) RefillRadio() error {
	status, err := c.Client().PlayerCurrentlyPlaying(c.Context)
	if err != nil {
		return err
	}
	paused := false
	if !status.Playing {
		paused = true
	}
	toRemove := []spotify.ID{}
	radioPlaylist, db, err := c.GetRadioPlaylist("")
	if err != nil {
		return err
	}

	playlistItems, err := c.Client().GetPlaylistItems(c.Context, radioPlaylist.ID)
	if err != nil {
		return fmt.Errorf("orig playlist items: %w", err)
	}

	if status.PlaybackContext.URI != radioPlaylist.URI || paused {
		return c.RadioFromPlaylist(radioPlaylist.SimplePlaylist)
	}

	page := 0
	for {
		tracks, err := c.Client().GetPlaylistItems(c.Context, radioPlaylist.ID, spotify.Limit(50), spotify.Offset(page*50))
		if err != nil {
			return fmt.Errorf("tracks: %w", err)
		}
		if len(tracks.Items) == 0 {
			break
		}
		for _, track := range tracks.Items {
			if track.Track.Track.ID == status.Item.ID {
				break
			}
			toRemove = append(toRemove, track.Track.Track.ID)
		}
		page++
	}
	if len(toRemove) > 0 {
		var trackGroups []spotify.ID
		for idx, item := range toRemove {
			if idx%100 == 0 && idx != 0 {
				_, err = c.Client().RemoveTracksFromPlaylist(c.Context, radioPlaylist.ID, trackGroups...)
				trackGroups = []spotify.ID{}
			}
			trackGroups = append(trackGroups, item)
			if err != nil {
				return fmt.Errorf("error clearing playlist: %w", err)
			}
		}
		_, err := c.Client().RemoveTracksFromPlaylist(c.Context, radioPlaylist.ID, trackGroups...)
		if err != nil {
			return err
		}
	}

	toAdd := 500 - (int(playlistItems.Total) - len(toRemove))
	playlistItems, err = c.Client().GetPlaylistItems(c.Context, radioPlaylist.ID)
	if err != nil {
		return fmt.Errorf("playlist items: %w", err)
	}
	total := playlistItems.Total
	pages := int(math.Ceil(float64(total) / 50))
	randomPage := 1
	if pages > 1 {
		randomPage = rand.Intn(pages-1) + 1
	}
	playlistPage, err := c.Client().
		GetPlaylistItems(c.Context, radioPlaylist.ID, spotify.Limit(50), spotify.Offset((randomPage-1)*50))
	if err != nil {
		return fmt.Errorf("playlist page: %w", err)
	}
	pageSongs := playlistPage.Items
	rand.Shuffle(len(pageSongs), func(i, j int) { pageSongs[i], pageSongs[j] = pageSongs[j], pageSongs[i] })
	seedCount := 5
	if len(pageSongs) < seedCount {
		seedCount = len(pageSongs)
	}
	seedIds := []spotify.ID{}
	for idx, song := range pageSongs {
		if idx >= seedCount {
			break
		}
		seedIds = append(seedIds, song.Track.Track.ID)
	}
	seed := spotify.Seeds{
		Tracks: seedIds,
	}
	recomendations, err := c.Client().GetRecommendations(c.Context, seed, &spotify.TrackAttributes{}, spotify.Limit(95))
	if err != nil {
		return err
	}
	recomendationIds := []spotify.ID{}
	for _, song := range recomendations.Tracks {
		exists, err := c.SongExists(db, song.ID)
		if err != nil {
			return fmt.Errorf("err check song existnce: %w", err)
		}
		if !exists {
			recomendationIds = append(recomendationIds, song.ID)
		}
	}
	queue := []spotify.ID{}
	for idx, rec := range recomendationIds {
		if idx > toAdd {
			break
		}
		_, err = db.QueryContext(c.Context, fmt.Sprintf("INSERT INTO radio (id) VALUES('%s')", rec.String()))
		if err != nil {
			return err
		}
		queue = append(queue, rec)
	}
	toAdd -= len(queue)
	_, err = c.Client().AddTracksToPlaylist(c.Context, radioPlaylist.ID, queue...)
	if err != nil {
		return fmt.Errorf("add tracks: %w", err)
	}
	err = c.Client().Repeat(c.Context, "context")
	if err != nil {
		return fmt.Errorf("repeat: %w", err)
	}
	for toAdd > 0 {
		id := rand.Intn(len(recomendationIds)-2) + 1
		seed := spotify.Seeds{
			Tracks: []spotify.ID{recomendationIds[id]},
		}
		additionalRecs, err := c.Client().
			GetRecommendations(c.Context, seed, &spotify.TrackAttributes{}, spotify.Limit(100))
		if err != nil {
			return fmt.Errorf("get recs: %w", err)
		}
		additionalRecsIds := []spotify.ID{}
		for idx, song := range additionalRecs.Tracks {
			exists, err := c.SongExists(db, song.ID)
			if err != nil {
				return fmt.Errorf("check song existence: %w", err)
			}
			if !exists {
				if idx > toAdd {
					break
				}
				additionalRecsIds = append(additionalRecsIds, song.ID)
				queue = append(queue, song.ID)
			}
		}
		toAdd -= len(queue)
		_, err = c.Client().AddTracksToPlaylist(c.Context, radioPlaylist.ID, additionalRecsIds...)
		if err != nil {
			return fmt.Errorf("add tracks to playlist: %w", err)
		}
	}
	return nil
}

func (c *Commander) RadioFromAlbum(album spotify.SimpleAlbum) error {
	tracks, err := c.AlbumTracks(album.ID, 1)
	if err != nil {
		return err
	}
	total := tracks.Total
	if total == 0 {
		return fmt.Errorf("this playlist is empty")
	}
	pages := int(math.Ceil(float64(total) / 50))
	randomPage := 1
	if pages > 1 {
		randomPage = rand.Intn(pages-1) + 1
	}
	albumTrackPage, err := c.AlbumTracks(album.ID, randomPage)
	if err != nil {
		return err
	}
	pageSongs := albumTrackPage.Tracks
	rand.Shuffle(len(pageSongs), func(i, j int) { pageSongs[i], pageSongs[j] = pageSongs[j], pageSongs[i] })
	seedCount := 5
	if len(pageSongs) < seedCount {
		seedCount = len(pageSongs)
	}
	seedIds := []spotify.ID{}
	for idx, song := range pageSongs {
		if idx >= seedCount {
			break
		}
		seedIds = append(seedIds, song.ID)
	}
	return c.RadioGivenList(seedIds[:seedCount], album.Name)
}

func (c *Commander) RadioGivenList(songs []spotify.ID, name string) error {
	seed := spotify.Seeds{
		Tracks: songs,
	}
	recomendations, err := c.Client().GetRecommendations(c.Context, seed, &spotify.TrackAttributes{}, spotify.Limit(99))
	if err != nil {
		return err
	}
	recomendationIds := []spotify.ID{}
	for _, song := range recomendations.Tracks {
		recomendationIds = append(recomendationIds, song.ID)
	}
	err = c.ClearRadio()
	if err != nil {
		return err
	}
	radioPlaylist, db, err := c.GetRadioPlaylist(name)
	if err != nil {
		return err
	}
	queue := []spotify.ID{songs[0]}
	for _, rec := range recomendationIds {
		exists, err := c.SongExists(db, rec)
		if err != nil {
			return err
		}
		if !exists {
			_, err := db.QueryContext(c.Context, fmt.Sprintf("INSERT INTO radio (id) VALUES('%s')", string(rec)))
			if err != nil {
				return err
			}
			queue = append(queue, rec)
		}
	}
	_, err = c.Client().AddTracksToPlaylist(c.Context, radioPlaylist.ID, queue...)
	if err != nil {
		return err
	}
	err = c.Client().PlayOpt(c.Context, &spotify.PlayOptions{
		PlaybackContext: &radioPlaylist.URI,
	})
	if err != nil {
		if isNoActiveError(err) {
			deviceID, err := c.activateDevice()
			if err != nil {
				return err
			}
			err = c.Client().PlayOpt(c.Context, &spotify.PlayOptions{
				PlaybackContext: &radioPlaylist.URI,
				DeviceID:        &deviceID,
			})
			if err != nil {
				return err
			}
		}
	}
	for i := 0; i < 4; i++ {
		id := rand.Intn(len(recomendationIds)-2) + 1
		seed := spotify.Seeds{
			Tracks: []spotify.ID{recomendationIds[id]},
		}
		additionalRecs, err := c.Client().GetRecommendations(c.Context, seed, &spotify.TrackAttributes{}, spotify.Limit(100))
		if err != nil {
			return err
		}
		additionalRecsIds := []spotify.ID{}
		for _, song := range additionalRecs.Tracks {
			exists, err := c.SongExists(db, song.ID)
			if err != nil {
				return err
			}
			if !exists {
				_, err = db.QueryContext(c.Context, fmt.Sprintf("INSERT INTO radio (id) VALUES('%s')", string(song.ID)))
				if err != nil {
					return err
				}
				additionalRecsIds = append(additionalRecsIds, song.ID)
			}
		}
		_, err = c.Client().AddTracksToPlaylist(c.Context, radioPlaylist.ID, additionalRecsIds...)
		if err != nil {
			return err
		}
	}
	return nil
}
