package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"

	"git.asdf.cafe/abs3nt/gspot/src/config"
)

var (
	auth         *spotifyauth.Authenticator
	ch           = make(chan *spotify.Client)
	state        = "abc123"
	configDir, _ = os.UserConfigDir()
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func GetClient(conf *config.Config) (c *spotify.Client, err error) {
	if conf.ClientId == "" || (conf.ClientSecret == "" && conf.ClientSecretCmd == "") || conf.Port == "" {
		return nil, fmt.Errorf("INVALID CONFIG")
	}
	if conf.ClientSecretCmd != "" {
		args := strings.Fields(conf.ClientSecretCmd)
		cmd := args[0]
		secret_command := exec.Command(cmd)
		if len(args) > 1 {
			secret_command.Args = args
		}
		secret, err := secret_command.Output()
		if err != nil {
			panic(err)
		}
		conf.ClientSecret = strings.TrimSpace(string(secret))
	}
	auth = spotifyauth.New(
		spotifyauth.WithClientID(conf.ClientId),
		spotifyauth.WithClientSecret(conf.ClientSecret),
		spotifyauth.WithRedirectURL(fmt.Sprintf("http://localhost:%s/callback", conf.Port)),
		spotifyauth.WithScopes(
			spotifyauth.ScopeImageUpload,
			spotifyauth.ScopePlaylistReadPrivate,
			spotifyauth.ScopePlaylistModifyPublic,
			spotifyauth.ScopePlaylistModifyPrivate,
			spotifyauth.ScopePlaylistReadCollaborative,
			spotifyauth.ScopeUserFollowModify,
			spotifyauth.ScopeUserFollowRead,
			spotifyauth.ScopeUserLibraryModify,
			spotifyauth.ScopeUserLibraryRead,
			spotifyauth.ScopeUserReadPrivate,
			spotifyauth.ScopeUserReadEmail,
			spotifyauth.ScopeUserReadCurrentlyPlaying,
			spotifyauth.ScopeUserReadPlaybackState,
			spotifyauth.ScopeUserModifyPlaybackState,
			spotifyauth.ScopeUserReadRecentlyPlayed,
			spotifyauth.ScopeUserTopRead,
			spotifyauth.ScopeStreaming,
		),
	)
	if _, err := os.Stat(filepath.Join(configDir, "gspot/auth.json")); err == nil {
		authFilePath := filepath.Join(configDir, "gspot/auth.json")
		authFile, err := os.Open(authFilePath)
		if err != nil {
			return nil, err
		}
		defer authFile.Close()
		tok := &oauth2.Token{}
		err = json.NewDecoder(authFile).Decode(tok)
		if err != nil {
			return nil, err
		}
		authCtx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{
			Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				slog.Debug("ROUND_TRIPPER", "request", r.URL.Path)
				return http.DefaultTransport.RoundTrip(r)
			}),
		})
		authClient := auth.Client(authCtx, tok)
		client := spotify.New(authClient)
		new_token, err := client.Token()
		if err != nil {
			return nil, err
		}
		out, err := json.MarshalIndent(new_token, "", " ")
		if err != nil {
			return nil, err
		}
		err = os.WriteFile(authFilePath, out, 0o600)
		if err != nil {
			return nil, fmt.Errorf("failed to save auth")
		}
		return client, nil
	}
	// first start an HTTP server
	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("AUTHENTICATOR", "received request", r.URL.String())
	})
	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", conf.Port),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		_ = server.ListenAndServe()
	}()
	url := auth.AuthURL(state)
	slog.Info("AUTH", "url", url)
	cmd := exec.Command("xdg-open", url)
	_ = cmd.Start()
	// wait for auth to complete
	client := <-ch

	_ = server.Shutdown(context.Background())
	// use the client to make calls that require authorization
	user, err := client.CurrentUser(context.Background())
	if err != nil {
		return nil, err
	}
	slog.Info("AUTH", "You are logged in as:", user.ID)
	return client, nil
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(r.Context(), state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		slog.Error("State mismatch: %s != %s\n", st, state)
		os.Exit(1)
	}
	out, err := json.MarshalIndent(tok, "", " ")
	if err != nil {
		slog.Error("AUTHENTICATOR", "failed to unmarshal", err)
		os.Exit(1)
	}
	err = os.WriteFile(filepath.Join(configDir, "gspot/auth.json"), out, 0o600)
	if err != nil {
		slog.Error("AUTHENTICATOR", "failed to save auth", err)
	}
	// use the token to get an authenticated client
	client := spotify.New(auth.Client(r.Context(), tok))
	fmt.Fprintf(w, "Login Completed!")
	ch <- client
}
