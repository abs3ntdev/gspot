package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	"connectrpc.com/connect"
	"go.uber.org/fx"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/abs3ntdev/gspot/src/config"

	"github.com/abs3ntdev/gspot/gen/gspot/v1/gspotv1connect"
)

// Run starts the daemon — writes PID file, starts HTTP server, blocks.
func Run(srv *Server, conf *config.Config, logger *slog.Logger) {
	if err := WritePid(conf); err != nil {
		logger.Error("Failed to write PID file", "error", err)
		os.Exit(1)
	}
	defer RemovePid(conf)

	if err := serveHTTP(srv, conf, logger); err != nil {
		logger.Error("Server error", "error", err)
		os.Exit(1)
	}
}

// NewPidFile registers fx lifecycle hooks to write/remove the PID file.
func NewPidFile(lc fx.Lifecycle, conf *config.Config, logger *slog.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Writing PID file", "path", conf.PidFile)
			return WritePid(conf)
		},
		OnStop: func(ctx context.Context) error {
			RemovePid(conf)
			return nil
		},
	})
}

// NewHTTPServer registers fx lifecycle hooks to start/stop the ConnectRPC HTTP server.
func NewHTTPServer(lc fx.Lifecycle, srv *Server, conf *config.Config, logger *slog.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := serveHTTP(srv, conf, logger); err != nil {
					logger.Error("Server error", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			os.Remove(conf.SocketPath)
			return nil
		},
	})
}

func serveHTTP(srv *Server, conf *config.Config, logger *slog.Logger) error {
	socketPath := conf.SocketPath

	if _, err := os.Stat(socketPath); err == nil {
		if err := os.Remove(socketPath); err != nil {
			return fmt.Errorf("failed to remove existing socket: %w", err)
		}
	}

	path, handler := gspotv1connect.NewGspotServiceHandler(
		srv,
		connect.WithInterceptors(loggingInterceptor(logger)),
	)

	mux := http.NewServeMux()
	mux.Handle(path, handler)

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

	logger.Info("Daemon is listening", "socket", socketPath)

	if err := httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("serve error: %w", err)
	}
	return nil
}

func loggingInterceptor(logger *slog.Logger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			logger.Debug("RPC request", "procedure", req.Spec().Procedure)
			resp, err := next(ctx, req)
			if err != nil {
				logger.Debug("RPC error", "procedure", req.Spec().Procedure, "error", err)
			}
			return resp, err
		}
	}
}
