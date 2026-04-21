<div align="center">

<h1>dockpose</h1>

<p>
  <a href="https://go.dev/">
    <img src="https://img.shields.io/badge/Go-1.25-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go 1.25">
  </a>
  <a href="https://github.com/charmbracelet/bubbletea">
    <img src="https://img.shields.io/badge/Bubble_Tea-TUI-FF7AC0?style=for-the-badge" alt="Bubble Tea TUI">
  </a>
  <a href="https://docs.docker.com/compose/">
    <img src="https://img.shields.io/badge/Docker_Compose-2496ed?style=for-the-badge&logo=docker&logoColor=white" alt="Docker Compose">
  </a>
  <a href="LICENSE">
    <img src="https://img.shields.io/badge/License-MIT-yellow?style=for-the-badge" alt="MIT License">
  </a>
  <img src="https://img.shields.io/badge/Status-pre--v0.1.0-f59e0b?style=for-the-badge" alt="pre-v0.1.0">
</p>

<br>

<img src="docs/media/demo.gif" alt="dockpose demo: stack list with status dots, dependency DAG, and live filter" width="100%">

<strong>k9s for Docker Compose stacks.</strong>
<br>
Auto-discovers your compose files, visualizes dependencies, manages stacks<br>as first-class units, and works over SSH — all from your keyboard.

<p>
  <a href="#try-it-without-docker">Try It</a>&nbsp;&nbsp;·&nbsp;&nbsp;<a href="#features">Features</a>&nbsp;&nbsp;·&nbsp;&nbsp;<a href="#how-it-works">How It Works</a>&nbsp;&nbsp;·&nbsp;&nbsp;<a href="#keybinds">Keybinds</a>&nbsp;&nbsp;·&nbsp;&nbsp;<a href="#install">Install</a>
</p>

</div>

## Why?

The dominant container TUI is container-centric by design. Most people running Compose on a homelab, a small fleet, or a single VPS think in **stacks**, not containers — and currently juggle them with `cd stack && docker compose up -d && cd ..` or a heavyweight web UI.

dockpose's bet: if the tool knows about your stacks, their dependencies, their profiles, their `.env` files, and their health, managing them becomes a 2-second keyboard operation instead of a 30-second shell dance.

> [!NOTE]
> dockpose is pre-v0.1.0. The TUI is functional and demo mode works end to end; cross-compiled releases and packaging (Homebrew/AUR/Scoop) are landing next.

## Try It Without Docker

```sh
dockpose --demo
```

Demo mode loads a synthetic homelab (media, monitoring, traefik, dev-api, authentik), rotates statuses on a 2s tick, and disables destructive actions so you can mash keys safely. **No Docker daemon required.**

## Features

<details open>
<summary><strong>Stack-First Design</strong></summary>

<br>

The primary list view is stacks, not containers. A container's identity is always "service X in stack Y." Stack-level up / down / restart / pull, with profile selection prompted when multiple profiles are defined.

</details>

<details open>
<summary><strong>Auto-Discovery</strong></summary>

<br>

Scans configured paths for `compose.y*ml` files on first launch and builds a registry cached at `~/.config/dockpose/stacks.toml`. Default scan paths cover `~/docker`, `~/homelab`, `~/projects`, `~/stacks`, and the current directory.

</details>

<details open>
<summary><strong>Dependency DAG Visualization</strong></summary>

<br>

Renders a layered ASCII-art dependency graph for each stack from its `depends_on` declarations. Per-service status dots (●/◐/○) make at-a-glance diagnosis trivial: green = healthy, yellow = unhealthy/starting, red = stopped.

</details>

<details>
<summary><strong>SSH-Native Remote Contexts</strong></summary>

<br>

Remote Docker contexts (`ssh://host`, `tcp://`, etc.) are first-class. Switch between local and remote daemons with a single keybind (`c`); the active context is always shown in the header chip.

</details>

<details>
<summary><strong>Streaming Log Viewer</strong></summary>

<br>

Tail logs for any service or for the whole stack. Toggleable follow, timestamps, line wrap, and a buffer search. Multi-service tail multiplexes streams with per-service prefixes.

</details>

<details>
<summary><strong>.env Editor with Secret Masking</strong></summary>

<br>

Edit your stack's `.env` from inside the TUI without leaking secrets to history. Values that look like secrets are masked by default; reveal them per-row with `r`.

</details>

<details>
<summary><strong>Demo Mode</strong></summary>

<br>

A synthetic `docker.Source` with five fixture stacks and rotating statuses, so you can demo, screenshot, or evaluate the tool without a Docker daemon. Destructive actions safely no-op with toast feedback.

</details>

<details>
<summary><strong>Single Static Binary</strong></summary>

<br>

One Go binary, no runtime, cross-compiled for macOS, Linux, and Windows on amd64 and arm64. Designed to land on Homebrew, AUR, Scoop, apt, and rpm.

</details>

## How It Works

### Architecture

dockpose is built on [Bubble Tea](https://github.com/charmbracelet/bubbletea) (Elm-style TUI) and [Lip Gloss](https://github.com/charmbracelet/lipgloss) for styling, with the official [Docker SDK for Go](https://github.com/docker/docker) for daemon communication. Compose parsing is done by [`compose-spec/compose-go`](https://github.com/compose-spec/compose-go) — the same library Compose itself uses, so v2 manifests parse identically to `docker compose`.

```
internal/
├── docker/      # Source interface; ClientSource (live) + container queries
├── stack/       # Compose parsing, dependency graph, registry, action shims
├── discover/    # Filesystem walker for compose.y*ml
├── demo/        # Synthetic Source: fixture stacks + rotating statuses
├── ui/          # Bubble Tea root + screens (stacklist, detail, logs, modals)
└── config/      # XDG config + cache I/O
```

### The Source Abstraction

The TUI talks to a single `docker.Source` interface — implemented by `ClientSource` (real Docker) and `demo.Source` (synthetic). This means demo mode and production mode share every line of UI code; bugs in either path surface in both.

### Stack-Aware Polling

A 2-second tick refreshes container states for all known stacks via the same query the detail view uses, so the dot strip on the list view and the per-service table never disagree. The detail view's `Refresh` is debounced to the same tick to avoid load on remote daemons over slow SSH links.

## Keybinds

<details open>
<summary><strong>Stack list</strong></summary>

<br>

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

</details>

<details>
<summary><strong>Stack detail</strong></summary>

<br>

| Key | Action |
| --- | --- |
| `l` | tail logs for the selected service |
| `s` | open shell in container |
| `x` | stop service |
| `R` | restart service |
| `i` | inspect container (JSON) |
| `esc` | back to stack list |

</details>

<details>
<summary><strong>Logs</strong></summary>

<br>

| Key | Action |
| --- | --- |
| `f` | toggle follow |
| `t` | toggle timestamps |
| `w` | toggle line wrap |
| `c` | clear buffer |
| `g` / `G` | top / bottom |
| `/` | filter buffer |

</details>

## Install

> [!IMPORTANT]
> Packaging (Homebrew, AUR, Scoop, apt, rpm) is on the v0.1.0 roadmap. For now, build from source.

```sh
git clone https://github.com/cwklurks/dockpose.git
cd dockpose
make check
./dockpose --version
```

**Requirements:** Go 1.25+, Docker, `make`, and `golangci-lint` for `make check`.

<details>
<summary><strong>Regenerating the demo recording</strong></summary>

<br>

Three paths, depending on what you have installed:

```sh
# 1. Cast file (zero system deps): records frames programmatically.
go run ./cmd/dockpose-record > docs/media/demo.cast

# 2. GIF from cast (requires `agg` from asciinema/agg releases):
agg --theme github-dark --font-size 13 --speed 1.2 \
    docs/media/demo.cast docs/media/demo.gif

# 3. Or, with vhs + ttyd installed, drive the real binary:
go build -o ./dockpose ./cmd/dockpose
vhs docs/media/demo.tape
```

</details>

## Roadmap

- **v0.1.0** — Cross-compiled release binaries via [GoReleaser](https://goreleaser.com), Homebrew tap, AUR PKGBUILD, Scoop manifest.
- **Persistent stack registry** — Honor the spec's `~/.config/dockpose/stacks.toml` cache and lazy refresh.
- **Filter persistence** — Remember active filter / cursor across sessions.
- **Pull progress** — Live `docker compose pull` progress in the status bar instead of a single toast.
- **Healthcheck-aware status** — Differentiate "running but unhealthy" via Docker's healthcheck output rather than treating all running containers as healthy.
- **Container events stream** — Replace polling with Docker's events API for sub-second updates.
- **Theming** — Multiple color themes (currently GitHub Dark only).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Bug reports and feature ideas welcome via [Issues](https://github.com/cwklurks/dockpose/issues).

## License

[MIT](LICENSE).

## Disclaimer

dockpose is an independent open-source project and is not affiliated with, endorsed by, or sponsored by Docker, Inc.
