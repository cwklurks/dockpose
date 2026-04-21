# dockpose

A keyboard-driven TUI for managing Docker Compose stacks.

> Status: pre-v0.1.0 - under active development.

![dockpose demo](docs/media/demo.gif)

## Try it without Docker

```sh
dockpose --demo
```

Demo mode loads a synthetic homelab (media, monitoring, traefik, dev-api,
authentik), rotates statuses on a 2s tick, and disables destructive actions
so you can mash keys safely. No Docker daemon required.

To regenerate the demo recording:

```sh
# Cast file (no extra deps): records frames programmatically.
go run ./cmd/dockpose-record > docs/media/demo.cast

# GIF (requires `agg` from asciinema/agg releases):
agg --theme github-dark --font-size 13 --speed 1.2 \
    docs/media/demo.cast docs/media/demo.gif

# Or, with `vhs + ttyd` installed, drive the real binary:
go build -o ./dockpose ./cmd/dockpose
vhs docs/media/demo.tape
```

## Why

Lazydocker is container-centric. Dockge is a web UI. Most people running Compose on a homelab or a small fleet want a keyboard-first, stack-aware, SSH-native tool. That's dockpose.

## Features (planned for v0.1.0)

- Auto-discovery of `compose.y*ml` files across configured paths
- Stack-level up / down / restart / pull
- Dependency DAG visualization from `depends_on`
- Profile picker for Compose profiles
- `.env` editor with secret masking
- Remote Docker contexts (SSH) as first-class citizens
- Log tailing with regex filter, follow toggle, timestamps
- Single static binary, cross-compiled

## Install

Packaging is planned for a later phase. For now, build from source.

## Build from source

Requirements: Go 1.25+, Docker, `make`, and `golangci-lint` for `make check`.

```sh
git clone https://github.com/cwklurks/dockpose.git
cd dockpose
make check
./dockpose --version
```

## Keybinds

**Stack list**

| Key | Action |
| --- | --- |
| `↑/k` `↓/j` | navigate |
| `g` / `G` | jump to top / bottom |
| `enter` | open stack detail |
| `u` | up (prompts for profile if multiple defined) |
| `d` | down |
| `r` | restart |
| `p` | pull |
| `e` | edit `.env` |
| `c` | switch Docker context |
| `/` | filter |
| `?` | help overlay |
| `q` | quit |

**Stack detail**

| Key | Action |
| --- | --- |
| `l` | tail logs for the selected service |
| `s` | open shell in container |
| `x` | stop service |
| `R` | restart service |
| `i` | inspect container (JSON) |
| `esc` | back to stack list |

## License

[MIT](LICENSE)

## Disclaimer

dockpose is an independent open-source project and is not affiliated with, endorsed by, or sponsored by Docker, Inc.
