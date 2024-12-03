package listenbrainz

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

type ListenBrainz struct {
	client *resty.Client

	labs *resty.Client
}

func NewListenBrainz(
	Endpoint string,
	ApiKey string,
) *ListenBrainz {
	c := resty.New().SetBaseURL(Endpoint)
	if ApiKey != "" {
		c = c.SetHeader("Authorization", "Token "+ApiKey)
	}
	return &ListenBrainz{
		client: c,
		labs:   resty.New().SetBaseURL("https://labs.api.listenbrainz.org"),
	}
}

type RadioApiResponse struct {
	Payload struct {
		Feedback []string `json:"feedback"`
		Jspf     struct {
			Playlist struct {
				Annotation string `json:"annotation"`
				Creator    string `json:"creator"`
				Extension  struct {
					HTTPSMusicbrainzOrgDocJspfPlaylist struct {
						Public bool `json:"public"`
					} `json:"https://musicbrainz.org/doc/jspf#playlist"`
				} `json:"extension"`
				Title string     `json:"title"`
				Track []ApiTrack `json:"track"`
			} `json:"playlist"`
		} `json:"jspf"`
	} `json:"payload"`
}

type ApiTrack struct {
	Album     string `json:"album"`
	Creator   string `json:"creator"`
	Duration  int    `json:"duration,omitempty"`
	Extension struct {
		HTTPSMusicbrainzOrgDocJspfTrack struct {
			ArtistIdentifiers []string `json:"artist_identifiers"`
			ReleaseIdentifier string   `json:"release_identifier"`
		} `json:"https://musicbrainz.org/doc/jspf#track"`
	} `json:"extension"`
	Identifier []string `json:"identifier"`
	Title      string   `json:"title"`
}

func (o *ListenBrainz) RequestRadio(ctx context.Context, req *RadioRequest) (*RadioTracksResponse, error) {
	var res RadioApiResponse
	resp, err := o.client.R().
		SetResult(&res).
		SetQueryParam("prompt", req.Prompt).
		SetQueryParam("mode", req.Prompt).
		Get("/1/explore/lb-radio")
	if err != nil {
		return nil, err
	}
	switch {
	case resp.StatusCode() == 200:
	default:
		return nil, fmt.Errorf("radio request code %d: %s", resp.StatusCode(), resp.Status())
	}
	tracks := res.Payload.Jspf.Playlist.Track
	return &RadioTracksResponse{
		Tracks: tracks,
	}, nil
}

type RadioTracksResponse struct {
	Tracks []ApiTrack
}

type RadioRequest struct {
	Prompt string
	Mode   string `json:"mode"`
}

type MatchTracksParams struct {
	Tracks []ApiTrack
}

func (o *ListenBrainz) MatchTracks(ctx context.Context, params *MatchTracksParams) error {
	// first try to get the mbid from the recording id
	o.labs.R().Get("/")
	return nil
}

type MatchedTrack struct {
	Mbid      string
	SpotifyId string
	Strategy  string
}

type TrackMatch struct {
	RecordingMbid   string   `json:"recording_mbid"`
	ArtistName      string   `json:"artist_name"`
	ReleaseName     string   `json:"release_name"`
	TrackName       string   `json:"track_name"`
	SpotifyTrackIds []string `json:"spotify_track_ids"`
}

var ErrNoMatch = fmt.Errorf("no match")

func (o *ListenBrainz) MatchTrack(ctx context.Context, track *ApiTrack) (*MatchedTrack, error) {
	// refuse to match a track with no identifiers
	if len(track.Identifier) == 0 {
		return nil, fmt.Errorf("%w: no identifier", ErrNoMatch)
	}
	var matches []TrackMatch
	// there are mbids, so try to get the first one that is valid
	resp, err := o.labs.R().
		SetResult(&matches).
		SetQueryParam("recording_mbid", track.Identifier[0]).
		Get("/spotify-id-from-mbid/json")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("labs request code %d: %s", resp.StatusCode(), resp.Status())
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("%w: no mbid", ErrNoMatch)
	}
	// for each match, see if ther eis a spotify id, and if there is, we are done!
	for _, match := range matches {
		if len(match.SpotifyTrackIds) == 0 {
			continue
		}
		return &MatchedTrack{
			Mbid:      match.RecordingMbid,
			SpotifyId: match.SpotifyTrackIds[0],
			Strategy:  "exact-match",
		}, nil
	}
	for _, match := range matches {
		var submatch []TrackMatch
		resp, err := o.labs.R().
			SetResult(&submatch).
			SetQueryParam("artist_name", match.ArtistName).
			SetQueryParam("release_name", match.ReleaseName).
			SetQueryParam("track_name", match.TrackName).
			Get("/spotify-id-from-track/json")
		if err != nil {
			return nil, err
		}
		if resp.StatusCode() != 200 {
			return nil, fmt.Errorf("labs request code %d: %s", resp.StatusCode(), resp.Status())
		}
		if len(submatch) == 0 {
			return nil, fmt.Errorf("%w: no tracks found", ErrNoMatch)
		}
		for _, submatch := range submatch {
			if len(submatch.SpotifyTrackIds) == 0 {
				continue
			}
			return &MatchedTrack{
				Mbid:      match.RecordingMbid,
				SpotifyId: submatch.SpotifyTrackIds[0],
				Strategy:  "track-match",
			}, nil
		}
	}

	return nil, fmt.Errorf("%w: no tracks found", ErrNoMatch)
}
