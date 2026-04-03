package cli

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"

	"github.com/abs3ntdev/gspot/src/components/daemon"
	"github.com/abs3ntdev/gspot/src/config"

	"github.com/abs3ntdev/gspot/gen/gspot/v1/gspotv1connect"
)

// newClient creates a ConnectRPC client that connects to the daemon over a Unix socket.
// If the daemon is not running, it automatically starts one.
func newClient(conf *config.Config) (gspotv1connect.GspotServiceClient, error) {
	// Ensure the daemon is running (auto-start if needed).
	if err := daemon.EnsureRunning(conf); err != nil {
		return nil, fmt.Errorf("failed to ensure daemon is running: %w", err)
	}

	// HTTP/2 transport over Unix socket (h2c — cleartext HTTP/2).
	transport := &http2.Transport{
		AllowHTTP: true,
		DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
			return net.Dial("unix", conf.SocketPath)
		},
	}

	httpClient := &http.Client{Transport: transport}

	// The base URL doesn't matter for Unix sockets — it just needs to be valid HTTP.
	client := gspotv1connect.NewGspotServiceClient(
		httpClient,
		"http://localhost",
		connect.WithGRPC(),
	)

	return client, nil
}
