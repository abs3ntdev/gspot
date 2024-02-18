package commands

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gfx.cafe/util/go/frand"
	"github.com/zmb3/spotify/v2"
	_ "modernc.org/sqlite"
)

func (c *Commander) Radio() error {
	current_song, err := c.Client.PlayerCurrentlyPlaying(c.Context)
	if err != nil {
		return err
	}
	if current_song.Item != nil {
		return c.RadioGivenSong(current_song.Item.SimpleTrack, current_song.Progress)
	}
	_, err = c.activateDevice(c.Context)
	if err != nil {
		return err
	}
	tracks, err := c.Client.CurrentUsersTracks(c.Context, spotify.Limit(10))
	if err != nil {
		return err
	}
	return c.RadioGivenSong(tracks.Tracks[frand.Intn(len(tracks.Tracks))].SimpleTrack, 0)
}

func (c *Commander) RadioGivenSong(song spotify.SimpleTrack, pos int) error {
	start := time.Now().UnixMilli()
	seed := spotify.Seeds{
		Tracks: []spotify.ID{song.ID},
	}
	recomendations, err := c.Client.GetRecommendations(c.Context, seed, &spotify.TrackAttributes{}, spotify.Limit(99))
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
	_, err = c.Client.AddTracksToPlaylist(c.Context, radioPlaylist.ID, queue...)
	if err != nil {
		return err
	}
	delay := time.Now().UnixMilli() - start
	if pos != 0 {
		pos = pos + int(delay)
	}
	err = c.Client.PlayOpt(c.Context, &spotify.PlayOptions{
		PlaybackContext: &radioPlaylist.URI,
		PositionMs:      pos,
	})
	if err != nil {
		if isNoActiveError(err) {
			deviceID, err := c.activateDevice(c.Context)
			if err != nil {
				return err
			}
			err = c.Client.PlayOpt(c.Context, &spotify.PlayOptions{
				PlaybackContext: &radioPlaylist.URI,
				DeviceID:        &deviceID,
				PositionMs:      pos,
			})
			if err != nil {
				return err
			}
		}
	}
	err = c.Client.Repeat(c.Context, "context")
	if err != nil {
		return err
	}
	for i := 0; i < 4; i++ {
		id := frand.Intn(len(recomendationIds)-2) + 1
		seed := spotify.Seeds{
			Tracks: []spotify.ID{recomendationIds[id]},
		}
		additional_recs, err := c.Client.GetRecommendations(c.Context, seed, &spotify.TrackAttributes{}, spotify.Limit(100))
		if err != nil {
			return err
		}
		additionalRecsIds := []spotify.ID{}
		for _, song := range additional_recs.Tracks {
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
		_, err = c.Client.AddTracksToPlaylist(c.Context, radioPlaylist.ID, additionalRecsIds...)
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
	err = c.Client.UnfollowPlaylist(c.Context, radioPlaylist.ID)
	if err != nil {
		return err
	}
	_, _ = db.Query("DROP TABLE IF EXISTS radio")
	configDir, _ := os.UserConfigDir()
	os.Remove(filepath.Join(configDir, "gospt/radio.json"))
	_ = c.Client.Pause(c.Context)
	return nil
}

func (c *Commander) GetRadioPlaylist(name string) (*spotify.FullPlaylist, *sql.DB, error) {
	configDir, _ := os.UserConfigDir()
	playlistFile, err := os.ReadFile(filepath.Join(configDir, "gospt/radio.json"))
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
	db, err := sql.Open("sqlite", filepath.Join(configDir, "gospt/radio.db"))
	if err != nil {
		return nil, nil, err
	}
	return playlist, db, nil
}

func (c *Commander) CreateRadioPlaylist(name string) (*spotify.FullPlaylist, *sql.DB, error) {
	// private flag doesnt work
	configDir, _ := os.UserConfigDir()
	playlist, err := c.Client.
		CreatePlaylistForUser(c.Context, c.User.ID, name+" - autoradio", "Automanaged radio playlist", false, false)
	if err != nil {
		return nil, nil, err
	}
	raw, err := json.MarshalIndent(playlist, "", " ")
	if err != nil {
		return nil, nil, err
	}
	err = os.WriteFile(filepath.Join(configDir, "gospt/radio.json"), raw, 0o600)
	if err != nil {
		return nil, nil, err
	}
	db, err := sql.Open("sqlite", filepath.Join(configDir, "gospt/radio.db"))
	if err != nil {
		return nil, nil, err
	}
	_, _ = db.QueryContext(c.Context, "DROP TABLE IF EXISTS radio")
	_, _ = db.QueryContext(c.Context, "CREATE TABLE IF NOT EXISTS radio (id string PRIMARY KEY)")
	return playlist, db, nil
}

func (c *Commander) SongExists(db *sql.DB, song spotify.ID) (bool, error) {
	song_id := string(song)
	sqlStmt := `SELECT id FROM radio WHERE id = ?`
	err := db.QueryRow(sqlStmt, song_id).Scan(&song_id)
	if err != nil {
		if err != sql.ErrNoRows {
			return false, err
		}

		return false, nil
	}

	return true, nil
}
