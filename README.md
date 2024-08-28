IF YOU ARE ON GITHUB.COM GO HERE INSTEAD: https://git.asdf.cafe/abs3nt/gspot

If you open an issue or PR on github I won't see it please use git.asdf.cafe. Register on asdf and open your PRs there

This project is still under heavy development and some things might not work or not work as intended. Don't hesitate to open an issue to let me know.

---

[![status-badge](https://ci.asdf.cafe/api/badges/abs3nt/gspot/status.svg)](https://ci.asdf.cafe/abs3nt/gspot)

# To install (with a package manager):

## Archlinux ([AUR])

`yay -S gspot-git`

# To build from source by pulling and building the binary

`git clone https://git.asdf.cafe/abs3nt/gspot`

`cd gspot`

`make build && sudo make install`

[AUR]: https://aur.archlinux.org/packages/gspot-git

# Configuration

go here https://developer.spotify.com/dashboard/applications to make a spotify application. you will need a client ID and a client secret. Set your redirect uri like this:

`http://localhost:8888/callback`

add your information to ~/.config/gspot/gspot.yml like this

```
client_id: "idgoeshere"
client_secret: "secretgoeshere"
port: "8888"
```

if you dont want to store your secret in the file in plaintext you can use a command to retreive it:

```
client_secret_cmd: "secret spotify_secret"
```

you should have either client_secret or client_secret_cmd

you can enable debug logging by adding

```
log_level: "debug"
log_output: "file"
```

it will log to ~/.config/gspot/gspot.log

## RUNNING

`gspot`

you will be asked to login, you will only have to do this the first time. After login you will be asked to select your default device.

helpful keybinds are shown in the bottom of the screen, hit ? to see all of them

To use the custom radio feature:

`gspot radio`

or hit ctrl+r on any track in the TUI. This will start an extended radio. To replenish the current radio run `gspot refillradio` and all the songs already listened will be removed and that number of new recomendations will be added.

This radio uses slightly different logic than the standard spotify radio to give a longer playlist and more recomendation. With a cronjob you can schedule refill to run to have an infinite and morphing radio station.

To view help:

`gspot --help`

Very open to contributations feel free to open a PR

[tmux plugin](https://git.asdf.cafe/abs3nt/tmux-gspot)

[wiki](https://git.asdf.cafe/abs3nt/gspot/wiki)
