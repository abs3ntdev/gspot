package daemon

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/abs3ntdev/gspot/src/components/commands"
	"github.com/abs3ntdev/gspot/src/config"

	"github.com/abs3ntdev/gspot/gen/gspot/v1/gspotv1connect"
)

// Run starts the daemon in the foreground. It writes a PID file, serves
// ConnectRPC over a Unix socket with h2c, and cleans up on SIGTERM/SIGINT.
func Run(c *commands.Commander, conf *config.Config, logger *slog.Logger) {
	if err := WritePid(conf); err != nil {
		log.Fatalf("Failed to write PID file: %v", err)
	}
	defer RemovePid(conf)

	for {
		err := startServer(c, conf, logger)
		if err != nil {
			log.Printf("Server error: %v", err)
			time.Sleep(time.Second)
			continue
		}
		break
	}
}

func startServer(c *commands.Commander, conf *config.Config, logger *slog.Logger) error {
	socketPath := conf.SocketPath

	if _, err := os.Stat(socketPath); err == nil {
		if err := os.Remove(socketPath); err != nil {
			return fmt.Errorf("failed to remove existing socket: %w", err)
		}
	}

	// Create ConnectRPC handler.
	server := NewServer(c)
	path, handler := gspotv1connect.NewGspotServiceHandler(
		server,
		connect.WithInterceptors(loggingInterceptor(logger)),
	)

	mux := http.NewServeMux()
	mux.Handle(path, handler)

	// Wrap with h2c for HTTP/2 cleartext (future bidi streaming support).
	h2cHandler := h2c.NewHandler(mux, &http2.Server{})

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("listen error: %w", err)
	}

	if err := os.Chmod(socketPath, 0o666); err != nil {
		listener.Close()
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	httpServer := &http.Server{Handler: h2cHandler}

	// Handle graceful shutdown on SIGTERM/SIGINT.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		logger.Info("Received signal, shutting down", "signal", sig)
		httpServer.Close()
		os.Remove(socketPath)
		RemovePid(conf)
		os.Exit(0)
	}()

	logger.Info("Daemon is listening", "socket", socketPath)

	if err := httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("serve error: %w", err)
	}
	return nil
}

// loggingInterceptor logs RPC requests at DEBUG level.
func loggingInterceptor(logger *slog.Logger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			logger.Debug("RPC request",
				"procedure", req.Spec().Procedure,
			)
			resp, err := next(ctx, req)
			if err != nil {
				logger.Debug("RPC error",
					"procedure", req.Spec().Procedure,
					"error", err,
				)
			}
			return resp, err
		}
	}
}
