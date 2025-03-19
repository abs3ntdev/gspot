package commands

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

	"git.asdf.cafe/abs3nt/gspot/src/listenbrainz"
	"github.com/zmb3/spotify/v2"
	_ "modernc.org/sqlite"
)

func (c *Commander) Radio() error {
	return nil
}

func (c *Commander) GetRecomendationIdsForPrompt(ctx context.Context, prompt string, mode string) ([]spotify.ID, error) {
	radioResp, err := c.lb.RequestRadio(ctx, &listenbrainz.RadioRequest{
		Prompt: prompt,
		Mode:   mode,
	})
	if err != nil {
		return nil, err
	}
	isMap := map[spotify.ID]struct{}{}
	for _, v := range radioResp.Tracks {
		match, err := c.lb.MatchTrack(ctx, &v)
		if err != nil {
			if errors.Is(err, listenbrainz.ErrNoMatch) {
				c.Log.Warn("no match for track", "err", err)
			} else {
				c.Log.Error("failed to match track", "err", err)
			}
			continue
		}
		if match.SpotifyId == "" {
			continue
		}
		isMap[spotify.ID(match.SpotifyId)] = struct{}{}
	}
	keys := make([]spotify.ID, 0, len(isMap))
	for i := range isMap {
		keys = append(keys, i)
	}
	return keys, nil
}

func (c *Commander) RadioFromPrompt(ctx context.Context, prompt string, mode string) error {
	list, err := c.GetRecomendationIdsForPrompt(ctx, prompt, mode)
	if err != nil {
		return err
	}
	err = c.RadioGivenList(list, fmt.Sprintf("lb - %s - %s", prompt, mode))
	if err != nil {
		return err
	}
	return nil
}

func (c *Commander) PlayRadio(radioPlaylist *spotify.FullPlaylist, pos int) error {
	err := c.Client().PlayOpt(c.Context, &spotify.PlayOptions{
		PlaybackContext: &radioPlaylist.URI,
		PositionMs:      spotify.Numeric(pos),
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
				PositionMs:      spotify.Numeric(pos),
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Commander) ClearRadio() error {
	radioPlaylist, err := c.GetRadioPlaylist("")
	if err != nil {
		return err
	}
	err = c.Client().UnfollowPlaylist(c.Context, radioPlaylist.ID)
	if err != nil {
		return err
	}
	_, _ = c.db.Query("DROP TABLE IF EXISTS radio")
	configDir, _ := os.UserConfigDir()
	os.Remove(filepath.Join(configDir, "gspot/radio.json"))
	_ = c.Client().Pause(c.Context)
	return nil
}

func (c *Commander) GetRadioPlaylist(name string) (*spotify.FullPlaylist, error) {
	configDir, _ := os.UserConfigDir()
	playlistFile, err := os.ReadFile(filepath.Join(configDir, "gspot/radio.json"))
	if errors.Is(err, os.ErrNotExist) {
		return c.CreateRadioPlaylist(name)
	}
	if err != nil {
		return nil, err
	}
	var playlist *spotify.FullPlaylist
	err = json.Unmarshal(playlistFile, &playlist)
	if err != nil {
		return nil, err
	}
	return playlist, nil
}

func (c *Commander) CreateRadioPlaylist(name string) (*spotify.FullPlaylist, error) {
	// private flag doesnt work
	configDir, _ := os.UserConfigDir()
	playlist, err := c.Client().
		CreatePlaylistForUser(c.Context, c.User.ID, name+" - autoradio", "Automanaged radio playlist", false, false)
	if err != nil {
		return nil, err
	}
	raw, err := json.MarshalIndent(playlist, "", " ")
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(filepath.Join(configDir, "gspot/radio.json"), raw, 0o600)
	if err != nil {
		return nil, err
	}
	_, _ = c.db.QueryContext(c.Context, "DROP TABLE IF EXISTS radio")
	_, _ = c.db.QueryContext(c.Context, "CREATE TABLE IF NOT EXISTS radio (id string PRIMARY KEY)")
	return playlist, nil
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

func (c *Commander) RadioGivenList(songs []spotify.ID, name string) error {
	err := c.ClearRadio()
	if err != nil {
		return err
	}
	radioPlaylist, err := c.GetRadioPlaylist(name)
	if err != nil {
		return err
	}
	queue := []spotify.ID{}
	for _, rec := range songs {
		exists, err := c.SongExists(c.db, rec)
		if err != nil {
			return err
		}
		if !exists {
			_, err := c.db.QueryContext(c.Context, fmt.Sprintf("INSERT INTO radio (id) VALUES('%s')", string(rec)))
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
	err = c.PlayRadio(radioPlaylist, 0)
	if err != nil {
		return err
	}
	for i := 0; i < 4; i++ {
		id := rand.Intn(len(songs)-2) + 1
		seed := spotify.Seeds{
			Tracks: []spotify.ID{songs[id]},
		}
		additionalRecs, err := c.Client().GetRecommendations(c.Context, seed, &spotify.TrackAttributes{}, spotify.Limit(100))
		if err != nil {
			return err
		}
		additionalRecsIds := []spotify.ID{}
		for _, song := range additionalRecs.Tracks {
			exists, err := c.SongExists(c.db, song.ID)
			if err != nil {
				return err
			}
			if !exists {
				_, err = c.db.QueryContext(c.Context, fmt.Sprintf("INSERT INTO radio (id) VALUES('%s')", string(song.ID)))
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
