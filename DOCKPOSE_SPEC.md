# dockpose — v1 Specification

**Tagline:** A keyboard-driven TUI for managing Docker Compose stacks.

**One-line pitch:** k9s for Compose stacks. Auto-discovers your compose files, visualizes dependencies, manages stacks as first-class units, works over SSH.

-----

## 1. Why this exists

The dominant container TUI (lazydocker) is container-centric by design. Its [issue #681](https://github.com/jesseduffield/lazydocker/issues/681) requesting stack-level up/down has been open for years with no plan to address it. Several small projects (DockTUI, Recomposable, ducker) have tried to fill the gap but none have nailed it. Meanwhile the homelab/self-hosted community runs 10–30 Compose stacks each and juggles them with `cd stack && docker compose up -d && cd ..` or a web UI like Dockge.

dockpose’s bet: the unit of mental work is the **stack**, not the container. If the tool knows about your stacks, their dependencies, their profiles, their .env files, and their health, managing them becomes a 2-second keyboard operation instead of a 30-second shell dance.

## 2. Target users

Three personas, in order of importance:

**The homelabber.** Runs 15 stacks across a NAS, a Pi, and a Proxmox VM. Has a Traefik stack, an Authelia stack, a media stack (Jellyfin, Sonarr, Radarr, Prowlarr), a monitoring stack, etc. Interacts with Docker primarily over SSH from a MacBook or Linux desktop. This user is terminal-native, opinionated, and will share the tool aggressively if it’s good.

**The small-team dev.** Works on a product with 4–8 stacks across dev/staging environments (app, db, redis, search, worker, queue). Uses Compose profiles to toggle dev vs. test scenarios. Currently uses lazydocker + terminal tabs. This user cares about profile switching, fast restarts, and per-service log tailing.

**The indie SaaS operator.** Runs a small fleet of Compose stacks on a single VPS. Has 2–5 projects, each a self-contained stack. Cares about `.env` editing, pulling new images, and seeing what’s wedged at a glance. SSH-heavy workflow.

What all three share: Compose is the substrate, stacks are the unit of work, the shell is home, and the browser is a tax.

## 3. Product principles

- **Stacks are first-class.** The primary list view is stacks, not containers. A container’s identity is always “service X in stack Y.”
- **Auto-discover everything.** First run should just work. Users shouldn’t have to register stacks.
- **Keyboard over menu.** Every action has a one or two-key binding. Menus exist for discovery, not daily use.
- **SSH-native.** Remote Docker contexts are a first-class citizen, not an afterthought. v1 supports them.
- **Don’t lie.** If the tool can’t do something safely (e.g., secrets handling), it says so. No silent plaintext writes, no “it worked” when it didn’t.
- **One binary, no runtime.** Single Go binary, cross-compiled, installable via Homebrew/AUR/Scoop/apt/rpm.

## 4. The v1 feature set

### 4.1 Stack discovery

On first launch, dockpose scans configured paths for `compose.y*ml` files and builds a stack registry cached at `~/.config/dockpose/stacks.toml`. Subsequent launches hydrate from cache and refresh lazily.

Default scan paths:

- `$XDG_CONFIG_HOME/dockpose/stacks.d/*.toml` (explicit registrations)
- `~/docker/*/compose.y*ml`
- `~/homelab/*/compose.y*ml`
- `~/projects/*/compose.y*ml`
- `~/stacks/*/compose.y*ml`
- `$PWD` (if launched from a stack dir, include it)

Scan depth defaults to 3 levels. Configurable in `~/.config/dockpose/config.toml`.

A stack is identified by its directory path + optional `name` field. If two compose files exist in the same directory (`compose.yml` + `compose.override.yml`), they’re treated as one stack with overrides.

### 4.2 Main view: the stack list

```
┌─ dockpose ─────────────────────────────── local ▼ ──── ? help ─┐
│                                                                │
│ STATUS  STACK              SERVICES   PROFILES    UPTIME      │
│ ●●●●●   media              5/5 up     default     3d 14h      │
│ ●●○○○   monitoring         2/4 up     default     2h 11m      │
│ ●       traefik            1/1 up     default     12d 3h      │
│ ○       dev-api            0/6 down   ---         ---         │
│ ●●●●    authentik          4/4 up     full        6d 22h      │
│                                                                │
│                                                                │
│ ↑↓ navigate  enter open  u up  d down  r restart  / filter    │
└────────────────────────────────────────────────────────────────┘
```

Columns:

- **Status:** One dot per service. Filled = healthy/running, half = starting/unhealthy, empty = stopped. Color-coded.
- **Stack:** Directory name or explicit name.
- **Services:** “running/total up” summary.
- **Profiles:** Active profile(s).
- **Uptime:** Oldest service’s uptime, or `---` if stopped.

Keybinds (top-level):

- `↑↓` / `j/k` — navigate
- `Enter` — open stack detail view
- `u` — up (prompts for profile if multiple defined)
- `d` — down (confirms)
- `r` — restart
- `p` — pull (async, status in footer)
- `l` — logs (multi-service tail)
- `e` — edit .env
- `/` — filter stacks
- `?` — help overlay
- `c` — switch Docker context (local / ssh://host / etc.)
- `q` — quit

### 4.3 Stack detail view

Opens when you hit Enter on a stack. Three panels:

**Top: service list with health.** Service name, container ID, image, status, ports, health. Same dot indicator as main view. Selectable.

**Middle: the dependency DAG.** This is the visual differentiator. Parses `depends_on` (both short form and long form with `condition`). Renders as a text graph with arrows. Nodes colored by current health. When you run `up`, the graph animates state transitions as services come online.

```
┌─ stack: media ─────────────────────────────────────────────────┐
│                                                                │
│   ┌─────────────┐                                              │
│   │ postgres  ● │──┐                                           │
│   └─────────────┘  │                                           │
│                    ▼                                           │
│   ┌─────────────┐  ┌─────────────┐   ┌─────────────┐          │
│   │ redis     ● │─▶│ sonarr    ● │──▶│ jellyfin  ● │          │
│   └─────────────┘  └─────────────┘   └─────────────┘          │
│          │                                                     │
│          ▼                                                     │
│   ┌─────────────┐                                              │
│   │ radarr    ● │                                              │
│   └─────────────┘                                              │
│                                                                │
│ ── services ────────────────────────────────────────────────── │
│ ● postgres       running   postgres:15        healthy   2d 3h │
│ ● redis          running   redis:7-alpine     healthy   2d 3h │
│ ● sonarr         running   linuxserver/sonarr healthy   2d 3h │
│ ● radarr         running   linuxserver/radarr healthy   2d 3h │
│ ● jellyfin       running   jellyfin/jellyfin  healthy   2d 2h │
│                                                                │
│ esc back  l logs  s shell  x stop  ↵ inspect service          │
└────────────────────────────────────────────────────────────────┘
```

**Bottom: footer with keybinds contextual to selected service.**

Keybinds in detail view:

- `l` — logs for selected service
- `s` — shell into selected service (`docker exec -it ... sh`, falls back through `bash`, `sh`, `ash`)
- `x` — stop selected service
- `R` — restart selected service
- `i` / `Enter` — inspect (json dump in a pager-style panel)
- `esc` — back to main view

### 4.4 Up with profile picker

Compose profiles are powerful but underused because activating them is friction-heavy (`--profile foo --profile bar`). dockpose surfaces them:

When you press `u` on a stack that declares profiles, a picker appears:

```
┌─ up: monitoring ─────────┐
│                          │
│ [x] default              │
│ [ ] alerting             │
│ [ ] debug                │
│ [ ] experimental         │
│                          │
│ space toggle  ↵ start    │
└──────────────────────────┘
```

Multi-select via space. Enter launches. Selection remembered per-stack across sessions.

### 4.5 .env editor with secret masking

Press `e` on a stack. Opens a modal editor showing the stack’s `.env` file with values masked by default.

```
┌─ .env: authentik ──────────────────────────────────────────────┐
│                                                                │
│   POSTGRES_PASSWORD     = ••••••••••••••              [show]  │
│   POSTGRES_USER         = authentik                           │
│   AUTHENTIK_SECRET_KEY  = ••••••••••••••••••••••     [show]  │
│   AUTHENTIK_ERROR_REPORT = true                               │
│   COMPOSE_PROJECT_NAME  = authentik                           │
│                                                                │
│                                                                │
│ ↑↓ navigate  ↵ edit  r reveal  s save  esc cancel             │
└────────────────────────────────────────────────────────────────┘
```

Rules:

- Values matching patterns like `*PASSWORD*`, `*SECRET*`, `*TOKEN*`, `*KEY*`, `*API*` are masked by default. Configurable.
- `r` reveals one value. `R` reveals all (requires confirmation).
- Saves atomically (write to `.env.tmp`, rename).
- Does not modify the compose file. Only `.env`.

Out of scope for v1: editing multi-line values, commented lines preservation (we preserve them but won’t let you edit comments), `.env.local` vs `.env` precedence handling.

### 4.6 Log tailing

Press `l` from anywhere to open log view:

- Main view: tails the full stack (`docker compose logs -f`).
- Detail view: tails the selected service.
- Multi-service mode: space-select services, then `l` to tail all selected with color-coded prefixes.

Log view features:

- `/` filter (regex)
- `g/G` top/bottom
- `f` follow toggle
- `w` word wrap toggle
- `t` timestamps toggle
- `c` clear screen (doesn’t affect buffer)
- `esc` back

### 4.7 Docker context switching

Press `c` from the main view. Picker shows contexts from `~/.docker/contexts/meta/`:

```
┌─ contexts ──────────────────┐
│                             │
│ ● local      (default)      │
│   nas        ssh://pi@nas   │
│   vps        ssh://root@vps │
│                             │
│ ↵ switch  esc cancel        │
└─────────────────────────────┘
```

Switching re-scans stacks under the new context. Stack paths on remote hosts are configured per-context in `config.toml`:

```toml
[contexts.nas]
scan_paths = ["/srv/docker", "/home/pi/stacks"]

[contexts.vps]
scan_paths = ["/opt/stacks"]
```

SSH connection reuses `docker context`’s existing config. No separate SSH setup.

### 4.8 What’s explicitly NOT in v1

Ship a focused v1. These are v2+ discussions:

- Editing compose files themselves
- Compose file syntax validation / linting
- Docker Swarm support
- Kubernetes support
- Portainer / Dockge import
- Image vulnerability scanning
- Resource graphs (CPU/mem sparklines)
- Web UI (there will never be a web UI; that’s Dockge’s job)
- Plugins / extension API
- Backup / snapshot management
- Compose file editor with autocomplete
- Per-stack custom actions (hooks)
- Notifications (email/discord/webhook)
- Multi-compose-file aggregation across directories

## 5. Technical architecture

### 5.1 Stack

- **Language:** Go 1.23+
- **TUI framework:** [Bubbletea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss) + [Bubbles](https://github.com/charmbracelet/bubbles)
- **Docker integration:** Official [Docker Go SDK](https://pkg.go.dev/github.com/docker/docker/client) + [compose-go](https://github.com/compose-spec/compose-go)
- **Config:** TOML via [BurntSushi/toml](https://github.com/BurntSushi/toml)
- **Distribution:** [goreleaser](https://goreleaser.com)

Rationale for Go + Bubbletea:

- Docker SDK is Go-native; Rust would mean shelling out to the `docker` CLI or maintaining a thin HTTP wrapper.
- Bubbletea’s Elm architecture (model → update → view) maps cleanly onto “stacks have state; events update them; views render them.”
- Single static binary, trivial cross-compilation, homelab distros (Alpine, Debian, Arch) all handle Go binaries without fuss.
- goroutines + channels are the natural fit for polling Docker events concurrently across many stacks without blocking the UI.

### 5.2 Module layout

```
dockpose/
├── cmd/
│   └── dockpose/
│       └── main.go              # entry, flag parsing, config load
├── internal/
│   ├── discover/                # compose file scanning
│   │   ├── discover.go
│   │   └── discover_test.go
│   ├── stack/                   # stack model, actions
│   │   ├── stack.go             # Stack type, StackState
│   │   ├── registry.go          # in-memory stack registry
│   │   ├── actions.go           # up/down/restart/pull/logs
│   │   └── deps.go              # DAG building from depends_on
│   ├── docker/                  # thin wrapper around Docker SDK
│   │   ├── client.go            # context-aware client factory
│   │   ├── events.go            # event stream multiplexer
│   │   └── logs.go              # log streaming
│   ├── dotenv/                  # .env parsing + writing
│   │   ├── parser.go            # preserves comments/order
│   │   ├── writer.go            # atomic writes
│   │   └── mask.go              # masking rules
│   ├── ui/                      # bubbletea models/views
│   │   ├── app.go               # top-level model
│   │   ├── stacklist/
│   │   ├── stackdetail/
│   │   ├── dag/                 # DAG rendering
│   │   ├── logs/
│   │   ├── envedit/
│   │   ├── profilepicker/
│   │   └── contextpicker/
│   └── config/                  # user config load/save
│       └── config.go
├── .goreleaser.yaml
├── go.mod
└── README.md
```

### 5.3 State model (Bubbletea)

Top-level `AppModel`:

```go
type AppModel struct {
    ctx        context.Context
    dockerCli  *docker.Client
    registry   *stack.Registry
    view       View              // stacklist | stackdetail | logs | envedit | ...
    current    *stack.Stack
    events     chan docker.Event // fan-in from all watched stacks
    cfg        *config.Config
    // view-specific sub-models
    stacklist  stacklist.Model
    detail     stackdetail.Model
    logs       logs.Model
    // ...
}
```

Events flow: Docker events channel → `tea.Cmd` → `Update(msg)` → new model → re-render.

The DAG view maintains its own layout cache keyed by stack ID + service graph hash. Invalidated only when the compose file changes on disk.

### 5.4 The DAG layout algorithm

This is the most technically interesting piece. Approach:

1. Parse `depends_on` from the normalized compose config (via `compose-go`). Build a directed graph where edges point from dependency to dependent (`db` → `app`).
1. Verify DAG-ness. Compose spec forbids cycles but verify anyway and gracefully fall back to a list view if violated.
1. Run a topological sort to compute layers. Nodes in the same layer have no dependency between them.
1. For each layer, compute x-positions with simple heuristic: minimize edge crossings by ordering nodes to match their dependency’s x-position (one pass, not optimal but fast enough for <30 services).
1. Render with ASCII box drawing. Arrows route through a simple grid. For stacks >15 services, collapse into a condensed vertical list view.

Speculation: the layout doesn’t need to be perfect. Even a simple Sugiyama-style layered layout will look dramatically better than anything competitors offer, which is currently “nothing.” Optimize later if users complain.

### 5.5 Concurrency model

- One long-running goroutine per active Docker context subscribes to `docker events` and forwards to a central channel.
- One goroutine per visible stack polls `docker compose ps` on a 2-second interval for status updates. Polling stops when the stack scrolls off-screen.
- Log tailing: one goroutine per tailed service, buffered channel to the UI, ring buffer (default 10k lines) to bound memory.
- All Docker SDK calls go through a `context.Context` with a 10-second timeout so a wedged daemon can’t freeze the UI.

### 5.6 Error handling philosophy

- Errors from Docker (daemon down, context unreachable, permission denied) are surfaced in a footer status bar, not modals. Users stay in flow.
- A `?` help screen has a section listing “recent errors” so nothing is lost.
- Destructive actions (down, stop, delete-volume-on-down) always confirm. No surprises.
- If a `.env` write fails, the editor stays open with the unsaved state and shows the error inline.

## 6. UX flow walkthroughs

### 6.1 First run

1. User installs via `brew install dockpose/tap/dockpose`.
1. Runs `dockpose`.
1. Tool detects no config, shows a one-shot onboarding screen:
- “Found 7 compose files in ~/docker, ~/homelab, ~/projects. Continue? [Y/n]”
- Writes default config.
1. Lands on main view with stacks discovered. First impression: “oh wow it just knew.”

### 6.2 Daily use: restart a wedged service

1. `dockpose` (already open in a tmux pane, always).
1. `/` filter “media”
1. Enter into stack detail.
1. `j` to sonarr.
1. `R` to restart.
1. DAG animates sonarr going yellow → green. Done in 4 keystrokes.

### 6.3 Bring up a stack with profiles

1. Main view, select “monitoring.”
1. `u`. Profile picker appears.
1. Space to toggle “default” and “alerting.”
1. Enter. Footer shows “starting monitoring (2 profiles)…” Stack appears with services coming online in DAG order.

### 6.4 Remote SSH workflow

1. `c` to open context picker.
1. Select `nas`.
1. Stacks rescan against the remote Docker daemon. Scan path is `/srv/docker` per context config.
1. Everything else works identically. Commands run on the remote host via `docker context use nas`.

## 7. Configuration reference

`~/.config/dockpose/config.toml`:

```toml
# Default scan paths (merged with built-in defaults)
scan_paths = [
    "~/docker",
    "~/homelab",
    "~/projects",
    "~/stacks",
]
scan_depth = 3

# Polling interval for stack status (seconds)
poll_interval = 2

# Log buffer size per service (lines)
log_buffer = 10000

# Patterns for auto-masking .env values (case-insensitive substring match)
mask_patterns = ["PASSWORD", "SECRET", "TOKEN", "KEY", "API", "CREDENTIAL"]

# Theme: "default", "dark", "light", "dracula", "nord", "catppuccin"
theme = "default"

# Confirm destructive actions
confirm_down = true
confirm_destroy = true

[contexts.local]
scan_paths = []  # falls back to top-level scan_paths

[contexts.nas]
scan_paths = ["/srv/docker", "/home/pi/stacks"]

[contexts.vps]
scan_paths = ["/opt/stacks"]
```

Stack-level overrides via `~/.config/dockpose/stacks.d/<name>.toml`:

```toml
name = "media"
path = "~/homelab/media"
display_name = "Media Stack"
pinned = true              # always top of list
auto_start = false         # exclude from bulk "up all"
```

## 8. Distribution plan

One-time setup via `goreleaser`. After that, every `git tag` produces releases across all channels automatically.

**Package managers:**

- Homebrew tap: `dockpose/homebrew-tap` → `brew install dockpose/tap/dockpose`
- AUR: `dockpose-bin` (binary) + `dockpose` (source build)
- Scoop bucket: `dockpose/scoop-bucket`
- Nix: `dockpose` in nixpkgs (submit after 500 stars)
- apt/rpm: goreleaser generates `.deb` and `.rpm`, hosted on GitHub releases
- `go install github.com/<user>/dockpose/cmd/dockpose@latest`

**Container image:** `ghcr.io/<user>/dockpose:latest` for users who want to run it in a container with the Docker socket mounted.

**One-line install:**

```sh
curl -sSL https://dockpose.dev/install.sh | sh
```

Shell script that detects OS/arch and pulls the right binary from GitHub releases.

## 9. Launch plan (week 5 of the project)

All passive channels, done once.

1. **Record a 90-second demo GIF with [VHS](https://github.com/charmbracelet/vhs).** Script: open → discover stacks → enter media stack → show DAG → restart sonarr → back out → switch context to nas → same flow on remote. This GIF goes at the top of the README.
1. **README structure:**
- GIF
- One-line pitch
- Install (one line per package manager)
- Keybind table
- Feature bullets with screenshots
- Comparison table vs. lazydocker / Dockge / Portainer
- “Not affiliated with Docker, Inc.” disclaimer
1. **Submissions:**
- PR to [awesome-tuis](https://github.com/rothgar/awesome-tuis)
- PR to [awesome-selfhosted](https://github.com/awesome-selfhosted/awesome-selfhosted)
- PR to [awesome-docker](https://github.com/veggiemonk/awesome-docker)
- Submit to [Terminal Trove](https://terminaltrove.com)
- Submit to Charm community showcase
1. **Launch posts:**
- Show HN: “Show HN: dockpose – a TUI for managing Docker Compose stacks”
- r/selfhosted (homelab-flavored body)
- r/homelab (cross-post)
- r/docker (cross-post)
- HN Tuesday–Thursday 9–11am ET, Reddit same day
1. **One blog post.** “Why I built dockpose.” Cross-posted to dev.to, Hashnode, personal blog. 1500 words max. Reference lazydocker #681, homelab experience, DAG feature.

## 10. Roadmap beyond v1

Not commitments, just a sketch so users can see the trajectory.

**v1.1 (post-launch, react to feedback):**

- Whatever the top 3 GitHub issues are
- Better error messages for common Docker daemon failures
- Keybind customization

**v1.2:**

- Resource graphs (CPU/mem sparklines) per service
- Image pull with inline progress bars
- Bulk operations: “up all,” “pull all,” “down all” with confirmation

**v1.3:**

- Compose file viewer (read-only, syntax-highlighted)
- Docker volume browser per stack
- Network inspection per stack

**v2:**

- Plugin system (Go plugins or Wasm) for custom actions
- Swarm support (maybe)
- Multi-host stack aggregation

Never:

- Web UI (Dockge owns that space)
- Kubernetes (k9s owns that space)
- GUI

## 11. Open questions / decisions to make early

Answer these in week 1 — they shape the code.

1. **How to handle stacks that change on disk while dockpose is running?** Options: fsnotify watcher, periodic rescan, manual refresh. Recommendation: fsnotify on compose files + manual refresh keybind (`F5`).
1. **What’s the abstraction boundary between compose CLI and Docker SDK?** Two approaches: shell out to `docker compose` for complex operations (up/down with profiles), use SDK for reads (ps/logs/inspect). Or: `compose-go` to parse + Docker SDK directly for everything, no shell-out. Second is faster and more robust but more code. Recommendation: hybrid — shell out for up/down (battle-tested), SDK for reads.
1. **How to test the UI?** Bubbletea has [teatest](https://github.com/charmbracelet/x/tree/main/exp/teatest) for model-level tests. Recommendation: teatest for model logic, manual testing for visuals, no visual regression tests in v1.
1. **Log line rendering performance for busy stacks.** A chatty Jellyfin + Sonarr + Radarr stack can emit 100+ lines/sec. Terminal rendering at that rate bottlenecks. Recommendation: coalesce log updates to 60fps (batch lines in a 16ms window), use `lipgloss.NewRenderer` with caching.
1. **How to handle Docker Compose v1 vs. v2?** v1 (`docker-compose`) is EOL but still exists on older distros. Recommendation: require v2, print a friendly error on v1 detection with install instructions.
1. **Color support detection.** Not every terminal supports truecolor. Recommendation: use `termenv` (ships with Bubbletea ecosystem) to auto-detect and degrade gracefully.

## 12. Success metrics (self-imposed)

- **Week 2 post-launch:** 500 GitHub stars, 10+ issues filed, 3+ contributors outside yourself
- **Month 2:** 2000 stars, packaged in Homebrew + AUR, a homelab YouTuber has featured it
- **Month 6:** 5000+ stars, steady issue flow, plugin system requests
- **Year 1:** it’s “the” Docker Compose TUI in people’s heads. Someone asks a Docker Compose question on r/selfhosted and a stranger replies “use dockpose”

-----

## Appendix A: Tech stack checklist

- [ ] Go 1.23+ installed
- [ ] Project scaffolded with the module layout above
- [ ] `bubbletea`, `lipgloss`, `bubbles` added via `go get`
- [ ] Docker SDK (`github.com/docker/docker`) and `compose-go` added
- [ ] `goreleaser init` run, `.goreleaser.yaml` configured for Darwin/Linux/Windows × amd64/arm64
- [ ] GitHub Actions workflow: `goreleaser release --snapshot` on PRs, `goreleaser release` on tags
- [ ] Pre-commit: `gofmt`, `go vet`, `golangci-lint`
- [ ] README with a VHS-generated GIF placeholder
- [ ] `LICENSE` (MIT recommended for max adoption)
- [ ] `CONTRIBUTING.md` with “how to dev locally” section
- [ ] `.editorconfig`

## Appendix B: Competitive landscape snapshot (as of April 2026)

Verify before launch — landscape moves.

|Tool        |Type|Scope                      |Status                                    |
|------------|----|---------------------------|------------------------------------------|
|lazydocker  |TUI |Container-centric          |Active, ~40k stars, will not add stack ops|
|ctop        |TUI |Container-centric, minimal |Stale (last release 2022)                 |
|ducker      |TUI |Container-centric, Ratatui |Active, smaller                           |
|DockTUI     |TUI |Stack-aware, Python/Textual|Early-stage                               |
|Recomposable|TUI |Stack-aware                |Early-stage                               |
|Dockge      |Web |Stack-aware                |Dominant web UI, 12k+ stars               |
|Portainer   |Web |Stack + containers + Swarm |Dominant enterprise                       |
|lazyjournal |TUI |Adjacent (logs only)       |Active                                    |

**dockpose’s wedge:** stack-first + DAG + SSH-native + keyboard-driven. None of the above combine all four.

## Appendix C: Naming, branding, legal

- **Name:** dockpose (confirmed available on GitHub)
- **Logo:** avoid Docker’s whale, avoid Moby. Speculation: a simple ASCII-art glyph or a cube-stack motif in a single accent color reads well on GitHub’s README.
- **Domain:** register `dockpose.dev` if available (~$15/yr), point at GitHub Pages or a one-page landing
- **Trademark:** “Docker” is a Docker Inc. trademark. Disclaimer in README: “dockpose is an independent open-source project and is not affiliated with, endorsed by, or sponsored by Docker, Inc.”
- **License:** MIT. Maximum permissive = maximum adoption for dev tools.
