# gspot — Agent Guide

A Spotify CLI controller and daemon built in Go. This document describes the architecture, conventions, and key decisions for AI agents working on this codebase.

## Architecture Overview

gspot is a single binary (`cmd/gspot/main.go`) that operates in two modes:

1. **CLI client** — sends commands to the daemon over a Unix socket via ConnectRPC
2. **Daemon** (`gspot daemon run`) — long-running process that holds the Spotify auth session and serves ConnectRPC over a Unix socket with h2c (HTTP/2 cleartext)

The daemon auto-starts when a CLI command is run and no daemon is running.

## Directory Structure

```
cmd/gspot/main.go              Entry point — just calls cli.Run()

proto/gspot/v1/gspot.proto      Protobuf service definition (source of truth for the RPC contract)
gen/gspot/v1/                   Generated protobuf + ConnectRPC stubs (committed to git)
buf.yaml / buf.gen.yaml         Buf codegen configuration

src/cli/                        CLI layer — owns the process, constructs fx containers per command
  cli.go                        Run(), global flags (--format, --json, --output), command assembly
  middleware.go                 rpcCommand() and command() wrappers, Service struct
  fx.go                         fx dependency sets (ConfigDeps, DaemonDeps), run() helper
  rpc.go                        newClient() — ConnectRPC client over Unix socket with auto-start
  output.go                     Format type, Output[T]() generic, CmdOutput(), JSON formatter
  pretty.go                     Per-response-type pretty printers
  playback.go                   play, pause, next, previous, volume, seek, shuffle, repeat, setdevice
  info.go                       nowplaying, status, devices, download_cover
  sharing.go                    link, linkcontext, youtube-link
  library.go                    like, unlike
  playlist.go                   playlists, playlist, play-playlist
  daemon.go                     daemon start/stop/status/run/restart

src/components/daemon/          Daemon internals
  daemon.go                     Run() — serves ConnectRPC over h2c Unix socket, signal handling, PID file
  server.go                     GspotServiceHandler implementation — delegates to Commander, maps to proto
  lifecycle.go                  PID file management, daemon start/stop/status, auto-start logic

src/components/commands/        Spotify API interaction layer (Commander)
  commander.go                  Commander struct — holds Spotify client, context, logger, cache
  errors.go                     IsNoActiveError(), IsRestrictionError()
  activate_device.go            ActivateDevice() — reads saved device from config
  play.go, pause.go, ...        Individual Spotify operations

src/components/cache/           Simple in-memory cache with TTL
src/components/logger/          slog logger setup
src/components/youtube/         YouTube link lookup (uses Google API)
src/config/                     Config struct with XDG-based defaults (uses github.com/adrg/xdg)
src/services/                   fx modules for config loading and Spotify OAuth
```

## Key Patterns

### CLI Command Pattern

Every CLI command follows this pattern:

```go
Action: rpcCommand(func(ctx context.Context, cmd *cli.Command, svc *Service) error {
    resp, err := svc.Client.SomeMethod(ctx, connect.NewRequest(&gspotv1.SomeRequest{...}))
    if err != nil {
        return err
    }
    return CmdOutput(cmd, resp.Msg, prettyPrinterFunc)
}),
```

- `rpcCommand()` builds an fx container with ConfigDeps, creates a ConnectRPC client (auto-starting daemon if needed), and passes `*Service` to the handler
- `CmdOutput()` reads `--format`/`--json`/`--output` flags and dispatches to the appropriate formatter
- Pretty printers are passed explicitly at the call site — no registry or type-switch

For commands that don't need the RPC client (daemon lifecycle), use `command()` directly with `fx.Invoke`.

### Output Formatting

```go
func Output[T proto.Message](format Format, w io.Writer, msg T, pretty func(io.Writer, T) error) error
func CmdOutput[T proto.Message](cmd *cli.Command, msg T, pretty func(io.Writer, T) error) error
```

- `pretty` is a type-specific formatter for human-readable output
- JSON uses `protojson` with `UseProtoNames: true` (snake_case fields)
- `FormatSilent` produces no output
- `prettyEmpty[T]()` is a generic no-op for commands with no meaningful response

### fx Dependency Injection

Two reusable dependency sets:

- `ConfigDeps` — config + logger (used by RPC client commands)
- `DaemonDeps` — full stack with cache, commander, managed context (used by `daemon run`)

`run(opts...)` builds a short-lived fx container (no lifecycle), executes Invoke functions, and returns. Used by all CLI commands except `daemon run`, which uses `fx.New(DaemonDeps, ...).Run()` for the full fx lifecycle.

### Server-Side Retry

Player commands (play, pause, shuffle, etc.) are wrapped with `retryPlayer()` which uses `avast/retry-go` to retry on Spotify's transient "Restriction violated" errors at 100ms, 250ms, 500ms intervals.

### Daemon Lifecycle

- **PID file**: `$XDG_RUNTIME_DIR/gspot/gspot.pid` (plain text, just the PID number)
- **Socket**: `$XDG_RUNTIME_DIR/gspot/gspot.sock` (Unix socket)
- Both paths are configurable via `~/.config/gspot/gspot.yml` (`pid_file`, `socket_path`)
- Falls back to `/tmp/gspot.pid` and `/tmp/gspot.sock` if `XDG_RUNTIME_DIR` is unset
- Auto-start: `newClient()` calls `daemon.EnsureRunning()` before connecting

## Adding a New Command

1. **Proto**: Add request/response messages and RPC method to `proto/gspot/v1/gspot.proto`
2. **Generate**: Run `make generate` (or `buf generate`)
3. **Server**: Implement the handler method on `*Server` in `src/components/daemon/server.go`
4. **Pretty printer**: Add a `prettyXxx` function in `src/cli/pretty.go`
5. **CLI command**: Add the command definition in the appropriate `src/cli/*.go` file using `rpcCommand()`
6. **Register**: Add to the `*Commands()` function in the category file, which is already wired into `allCommands()` in `cli.go`

## Build & Run

```sh
make generate    # regenerate proto stubs (only needed after proto changes)
make build       # build the binary to dist/
make daemon      # build and run daemon in foreground
make tidy        # go mod tidy
make install     # install to /usr/bin
```

## Config

Config file: `~/.config/gspot/gspot.yml`

```yaml
client_id: "your_spotify_client_id"
client_secret: "your_spotify_client_secret"
port: "8888"
log_level: "info"        # debug, info, warn, error
log_output: "stdout"
socket_path: ""          # default: $XDG_RUNTIME_DIR/gspot/gspot.sock
pid_file: ""             # default: $XDG_RUNTIME_DIR/gspot/gspot.pid
```

## Dependencies

Key dependencies:
- `connectrpc.com/connect` — RPC framework (over standard net/http)
- `google.golang.org/protobuf` — protobuf runtime
- `github.com/zmb3/spotify/v2` — Spotify Web API client
- `github.com/urfave/cli/v3` — CLI framework
- `go.uber.org/fx` — dependency injection
- `github.com/adrg/xdg` — XDG base directory paths
- `github.com/avast/retry-go/v4` — retry with backoff
- `golang.org/x/net/http2/h2c` — HTTP/2 cleartext for Unix socket
- `github.com/abs3ntdev/gunner` — config loading

## Important Notes

- **Generated code is committed** — no `buf` required to build, only to regenerate after proto changes
- **Commander is a separate layer** — the ConnectRPC server wraps Commander methods, it does not expose Commander directly. Commander stays focused on Spotify API interaction.
- **No tests currently** — the codebase has zero test files
- **YouTube link feature** uses Google API with separate OAuth (`client_secret.json`) — needs cleanup
- **OAuth state is hardcoded** (`"abc123"` in `src/services/auth.go`) — security issue for production use
