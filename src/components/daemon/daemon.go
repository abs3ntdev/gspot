package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"connectrpc.com/connect"
	"go.uber.org/fx"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/abs3ntdev/gspot/src/components/commands"
	"github.com/abs3ntdev/gspot/src/config"

	"github.com/abs3ntdev/gspot/gen/gspot/v1/gspotv1connect"
)

// Run registers the daemon server with the fx lifecycle. The server starts
// on OnStart and shuts down gracefully on OnStop (triggered by SIGTERM/SIGINT
// via fx's built-in signal handling).
func Run(lc fx.Lifecycle, c *commands.Commander, conf *config.Config, logger *slog.Logger) error {
	socketPath := conf.SocketPath

	// Clean up stale socket if present.
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
	httpServer := &http.Server{Handler: h2cHandler}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := WritePid(conf); err != nil {
				return fmt.Errorf("failed to write PID file: %w", err)
			}

			listener, err := net.Listen("unix", socketPath)
			if err != nil {
				RemovePid(conf)
				return fmt.Errorf("listen error: %w", err)
			}

			if err := os.Chmod(socketPath, 0o600); err != nil {
				listener.Close()
				RemovePid(conf)
				return fmt.Errorf("failed to set socket permissions: %w", err)
			}

			logger.Info("Daemon is listening", "socket", socketPath)

			// Serve in background so fx startup can complete.
			go func() {
				if err := httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
					logger.Error("Server error", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Shutting down daemon")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			err := httpServer.Shutdown(shutdownCtx)
			os.Remove(socketPath)
			RemovePid(conf)
			return err
		},
	})
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
