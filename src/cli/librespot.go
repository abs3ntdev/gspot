package cli

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"os/exec"
	"strings"
	"time"

	golibrespot "github.com/devgianlu/go-librespot"
	"github.com/devgianlu/go-librespot/apresolve"
	"github.com/devgianlu/go-librespot/session"
	"github.com/urfave/cli/v3"
	"go.uber.org/fx"
	"golang.org/x/oauth2"
	spotifyoauth2 "golang.org/x/oauth2/spotify"

	librespotpkg "github.com/abs3ntdev/gspot/src/components/librespot"
	"github.com/abs3ntdev/gspot/src/config"
)

func librespotCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:     "librespot",
			Usage:    "Manage the librespot player",
			Category: "Librespot",
			Commands: []*cli.Command{
				{
					Name:  "auth",
					Usage: "Authenticate with Spotify for librespot playback",
					Flags: []cli.Flag{
						&cli.IntFlag{
							Name:  "port",
							Usage: "Callback port for OAuth",
							Value: 0,
						},
					},
					Action: func(ctx context.Context, cmd *cli.Command) error {
						if cmd.Args().Present() {
							return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
						}
						return run(
							ConfigDeps,
							fx.Invoke(func(conf *config.Config, log *slog.Logger) error {
								return librespotAuth(conf, log, int(cmd.Int("port")))
							}),
						)
					},
				},
				{
					Name:  "status",
					Usage: "Check librespot player status",
					Action: func(ctx context.Context, cmd *cli.Command) error {
						if cmd.Args().Present() {
							return fmt.Errorf("unexpected arguments: %s", strings.Join(cmd.Args().Slice(), " "))
						}
						return run(
							ConfigDeps,
							fx.Invoke(func(conf *config.Config, log *slog.Logger) error {
								return librespotStatus(conf, log)
							}),
						)
					},
				},
			},
		},
	}
}

func librespotAuth(conf *config.Config, log *slog.Logger, callbackPort int) error {
	configDir := config.ConfigDir()
	llog := librespotpkg.NewSlogAdapter(log)

	// Load or create app state
	appState := &golibrespot.AppState{}
	appState.SetLogger(llog)
	_ = appState.Read(configDir)

	// Generate device ID if needed
	if appState.DeviceId == "" {
		deviceIdBytes := make([]byte, 20)
		_, _ = rand.Read(deviceIdBytes)
		appState.DeviceId = hex.EncodeToString(deviceIdBytes)
	}

	// Run the OAuth flow ourselves so we control the browser opening
	ctx := context.Background()
	serverCtx, serverCancel := context.WithCancel(ctx)
	defer serverCancel()

	port, codeCh, err := session.NewOAuth2Server(serverCtx, llog, callbackPort)
	if err != nil {
		return fmt.Errorf("failed starting OAuth server: %w", err)
	}

	oauthConf := &oauth2.Config{
		ClientID:    golibrespot.ClientIdHex,
		RedirectURL: fmt.Sprintf("http://127.0.0.1:%d/login", port),
		Scopes: []string{
			"app-remote-control", "playlist-modify", "playlist-modify-private",
			"playlist-modify-public", "playlist-read", "playlist-read-collaborative",
			"playlist-read-private", "streaming", "ugc-image-upload",
			"user-follow-modify", "user-follow-read", "user-library-modify",
			"user-library-read", "user-modify", "user-modify-playback-state",
			"user-modify-private", "user-personalized", "user-read-birthdate",
			"user-read-currently-playing", "user-read-email", "user-read-play-history",
			"user-read-playback-position", "user-read-playback-state",
			"user-read-private", "user-read-recently-played", "user-top-read",
		},
		Endpoint: spotifyoauth2.Endpoint,
	}

	verifier := oauth2.GenerateVerifier()
	authURL := oauthConf.AuthCodeURL("", oauth2.S256ChallengeOption(verifier))

	// Open browser
	fmt.Println("Opening browser for Spotify authentication...")
	if err := exec.Command("xdg-open", authURL).Start(); err != nil {
		// If xdg-open fails, print the URL for manual opening
		fmt.Printf("Could not open browser. Visit this URL:\n%s\n", authURL)
	}

	// Wait for callback
	fmt.Println("Waiting for authentication...")
	code := <-codeCh
	serverCancel()

	// Exchange code for token
	token, err := oauthConf.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return fmt.Errorf("failed exchanging OAuth code: %w", err)
	}

	username, ok := token.Extra("username").(string)
	if !ok || username == "" {
		return fmt.Errorf("no username in token response")
	}

	// Create session with the token to get stored credentials
	httpClient := &http.Client{Timeout: 30 * time.Second}
	resolver := apresolve.NewApResolver(llog, httpClient)

	sess, err := session.NewSessionFromOptions(ctx, &session.Options{
		Log:        llog,
		DeviceType: 1, // COMPUTER
		DeviceId:   appState.DeviceId,
		Credentials: session.SpotifyTokenCredentials{
			Username: username,
			Token:    token.AccessToken,
		},
		Resolver: resolver,
		Client:   httpClient,
		AppState: appState,
	})
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Store credentials
	appState.Credentials.Username = sess.Username()
	appState.Credentials.Data = sess.StoredCredentials()

	if err := appState.Write(); err != nil {
		sess.Close()
		return fmt.Errorf("failed saving credentials: %w", err)
	}

	sess.Close()

	fmt.Printf("Authenticated as %s\n", sess.Username())
	fmt.Println("Credentials saved. The daemon will use them on next start.")
	return nil
}

func librespotStatus(conf *config.Config, log *slog.Logger) error {
	if !conf.Librespot.Enabled {
		fmt.Println("Librespot is not enabled in config")
		fmt.Println("Set librespot.enabled: true in ~/.config/gspot/gspot.yml")
		return nil
	}

	configDir := config.ConfigDir()
	appState := &golibrespot.AppState{}
	appState.SetLogger(librespotpkg.NewSlogAdapter(log))
	_ = appState.Read(configDir)

	fmt.Printf("Librespot: enabled\n")
	fmt.Printf("Device name: %s\n", conf.Librespot.DeviceName)
	fmt.Printf("Audio backend: %s\n", conf.Librespot.AudioBackend)
	fmt.Printf("Bitrate: %d kbps\n", conf.Librespot.Bitrate)
	fmt.Printf("Zeroconf: %v\n", conf.Librespot.Zeroconf)

	if appState.Credentials.Username != "" {
		fmt.Printf("Authenticated as: %s\n", appState.Credentials.Username)
	} else {
		fmt.Println("Not authenticated. Run 'gspot librespot auth' to log in.")
	}

	return nil
}
