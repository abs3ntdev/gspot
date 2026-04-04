# gspot

A Spotify CLI controller and daemon built in Go. Single binary, no TUI -- just fast commands and a background daemon communicating over ConnectRPC on a Unix socket.

> This project is under active development. Don't hesitate to open an issue if something doesn't work as expected.

## Installation

### Arch Linux ([AUR](https://aur.archlinux.org/packages/gspot-git))

```sh
yay -S gspot-git
```

### From source

```sh
git clone https://github.com/abs3ntdev/gspot
cd gspot
make build && sudo make install
```

### From releases

Pre-built binaries for Linux, macOS, and Windows are available on the [releases page](https://github.com/abs3ntdev/gspot/releases).

## Configuration

1. Create a Spotify application at https://developer.spotify.com/dashboard/applications
2. Set the redirect URI to `http://127.0.0.1:8888/callback`
3. Create `~/.config/gspot/gspot.yml`:

```yaml
client_id: "your_client_id"
client_secret: "your_client_secret"
port: "8888"
```

If you don't want to store your secret in plaintext, use a command:

```yaml
client_secret_cmd: "secret spotify_secret"
```

You should have either `client_secret` or `client_secret_cmd`, not both.

### Optional settings

```yaml
log_level: "info"      # debug, info, warn, error
log_output: "stdout"   # stdout or file (~/.config/gspot/gspot.log)

# Override default paths (defaults use $XDG_RUNTIME_DIR/gspot/)
socket_path: "/path/to/gspot.sock"
pid_file: "/path/to/gspot.pid"
```

## Usage

On first run you'll be prompted to log in to Spotify via your browser. This only happens once.

The daemon starts automatically when you run any command. You can also manage it explicitly with `gspot daemon`.

### Output formats

All commands support `--format` (`pretty`, `json`, `silent`), `--json` (shorthand for `--format=json`), and `--output <file>` to write output to a file.

### Commands

#### Playback

| Command | Aliases | Description |
|---|---|---|
| `gspot play` | `pl`, `start`, `s` | Resume playback |
| `gspot playurl <url>` | `plu` | Play a Spotify URL |
| `gspot pause` | `pa` | Pause playback |
| `gspot toggleplay` | `t` | Toggle play/pause |
| `gspot next [amount]` | `n`, `skip` | Skip to the next track (optionally skip multiple) |
| `gspot previous` | `b`, `prev`, `back` | Go to the previous track |
| `gspot repeat` | | Toggle repeat mode |
| `gspot shuffle` | | Toggle shuffle mode |
| `gspot setdevice <device_id>` | | Set the active playback device |
| `gspot seek <position_ms>` | `sk` | Seek to a position (milliseconds) |
| `gspot seek forward` | `sk f` | Seek forward |
| `gspot seek backward` | `sk b` | Seek backward |
| `gspot volume up <percent>` | `v up` | Increase volume |
| `gspot volume down <percent>` | `v down`, `v dn` | Decrease volume |
| `gspot volume mute` | `v m` | Mute |
| `gspot volume unmute` | `v um` | Unmute |
| `gspot volume togglemute` | `v tm` | Toggle mute |

#### Info

| Command | Aliases | Description |
|---|---|---|
| `gspot nowplaying` | `now` | Print the current track (`--force` / `-f` to bypass cache) |
| `gspot status` | | Print full player status |
| `gspot devices` | `d` | List available devices |
| `gspot download_cover [path]` | `dl` | Download the current track's cover art |

#### Sharing

| Command | Aliases | Description |
|---|---|---|
| `gspot link` | `yy` | Print the current track's Spotify link |
| `gspot linkcontext` | `lc` | Print the current album/playlist link |
| `gspot youtube-link` | `yl` | Print the current track's YouTube link |

#### Library

| Command | Aliases | Description |
|---|---|---|
| `gspot like` | `l` | Like the current track |
| `gspot unlike` | `u` | Unlike the current track |

#### Playlists

| Command | Aliases | Description |
|---|---|---|
| `gspot playlists` | `pls` | List your playlists (`--limit`, `--offset`) |
| `gspot playlist <id>` | `pl-info` | Show tracks in a playlist |
| `gspot play-playlist <id> [offset]` | `plp` | Play a playlist (optionally from a track offset) |

#### Daemon

| Command | Description |
|---|---|
| `gspot daemon start` | Start the daemon in the background |
| `gspot daemon stop` | Stop the running daemon |
| `gspot daemon status` | Check if the daemon is running |
| `gspot daemon run` | Run the daemon in the foreground |
| `gspot daemon restart` | Restart the daemon |

## Architecture

gspot is a single binary that operates in two modes:

- **CLI client** -- sends commands to the daemon over a Unix socket via ConnectRPC
- **Daemon** (`gspot daemon run`) -- long-running process that holds the Spotify auth session and serves ConnectRPC over a Unix socket with h2c

The daemon auto-starts when a CLI command is run and no daemon is detected. Communication uses protocol buffers defined in `proto/gspot/v1/gspot.proto`.

## Integrations

- [tmux-gspot](https://github.com/abs3ntdev/tmux-gspot) -- tmux status bar plugin

The daemon's Unix socket interface makes it easy to build custom integrations and scripts. Use `--json` for machine-readable output.

## Build targets

```sh
make generate    # regenerate proto stubs (requires buf)
make build       # build to dist/
make run         # build and run
make daemon      # build and run daemon in foreground
make tidy        # go mod tidy
make install     # install to /usr/bin (with shell completions)
make uninstall   # remove from /usr/bin
make clean       # remove dist/
```

## Contributing

Contributions are welcome -- feel free to open a PR.

## License

[MIT](LICENSE)
