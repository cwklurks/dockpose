# Changelog

All notable changes to dockpose will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned

- AUR `dockpose-bin` and Scoop manifest via GoReleaser
- Persistent stack registry cache at `~/.config/dockpose/stacks.toml`
- Healthcheck-aware status (read Docker `State.Health.Status` rather than treating every running container as healthy)
- Container events stream (replace 2s polling with the Docker events API for sub-second updates)
- Additional color themes (currently GitHub Dark only)

## [0.1.6] — 2026-04-21

Distribution-only release. The `dockpose` binary is byte-identical to 0.1.0; this is purely about getting it onto your machine more easily.

### Added

- **Homebrew tap.** `cwklurks/tap/dockpose` Formula pushed automatically from GoReleaser on every future tag. After `brew tap cwklurks/tap`, install with the bare `brew install dockpose`.
- **README install section.** Explicit paths for Homebrew, raw `curl` to tarball, `.deb`, `.rpm`, Windows zip, and build-from-source.

### Fixed

- **Release pipeline plumbing.** Several iterative CI fixes (`v0.1.1` through `v0.1.5`, now deleted) landed between 0.1.0 and 0.1.6: the GoReleaser action's `args` input, exposing `HOMEBREW_TAP_GITHUB_TOKEN` to the release step, switching from an invalid `envOrDefault` template to a direct `.Env` reference, seeding the empty homebrew-tap repo with a `main` branch, and finally correcting the PAT scope so GoReleaser could `PUT` a Formula.

## [0.1.0] — 2026-04-21

Initial public release. Ships cross-compiled binaries for Linux, macOS, and Windows on both amd64 and arm64, plus `.deb` and `.rpm` packages.

### Added

- **Auto-discovery** of `compose.y*ml` under configured scan paths (defaults: `~/docker`, `~/homelab`, `~/projects`, `~/stacks`, and the current directory).
- **Stack list view** with a colored status-dot strip per stack (● healthy / ◐ starting/unhealthy / ○ stopped), services-up / total count, profiles, and path. Columns match the tagline in the spec.
- **Stack detail view** with:
  - An ASCII-art **dependency DAG** rendered from `depends_on` declarations, color-coded by per-service status.
  - A **service table** (name / image / status / health / uptime / ports), navigable with `j`/`k`.
- **Stack-level actions.** Single-key bindings for up (`u`), down (`d`), restart (`r`), pull (`p`), and per-service stop (`x`), restart (`R`), shell (`s`), inspect (`i`), logs (`l`).
- **Profile picker modal.** When a stack defines more than one Compose profile, `u` prompts for which to activate instead of silently using `default`.
- **Streaming log viewer.** Per-service or multi-service tails with toggles for follow (`f`), timestamps (`t`), line wrap (`w`), clear (`c`), and a filter (`/`). Multi-service tails prefix each line with the service name.
- **.env editor modal** with secret masking. Values that look like tokens or keys are masked by default; reveal per-row with `r`.
- **Docker context switcher.** `c` lists contexts from `docker context ls` and switches the active one — remote SSH contexts are first-class.
- **Filter** (`/`) and **help overlay** (`?`) on the stack list.
- **Responsive layout.** Tracks `tea.WindowSizeMsg` and resizes panels to fit. Alt-screen mode keeps the host terminal scrollback clean.
- **Demo mode** (`--demo`). A synthetic `docker.Source` implementation backed by five fixture stacks (media, monitoring, traefik, dev-api, authentik) with statuses that rotate on the 2s polling tick. Destructive actions (up/down/restart/pull) no-op with a toast so the UI can be safely explored without a Docker daemon. `dockpose --demo` is the fastest way to evaluate the tool.
- **`cmd/dockpose-record`**, a standalone utility that scripts a demo scenario and writes an asciinema v2 cast to stdout. Used to generate `docs/media/demo.gif` without needing `ttyd` or any external recorder.
- **Docker SDK integration** via `docker/docker` for container listing and log streaming; Compose parsing via `compose-spec/compose-go/v2` so v2 manifests parse identically to `docker compose` itself.

### Architecture highlights

- **`docker.Source` interface** abstracts container-state retrieval so demo mode and real mode share every line of UI code. `ClientSource` wraps the Docker SDK; `demo.Source` produces synthetic `ContainerState` slices.
- **2-second polling loop** refreshes both the stack list's aggregate status and the detail view's per-service state; the detail view's `Refresh` is debounced on the same tick to avoid thrashing slow SSH links.

### Known gaps

- Every running container is reported as healthy (no healthcheck introspection yet).
- Stack discovery rescans on every launch instead of reading the `~/.config/dockpose/stacks.toml` cache the spec envisions.
- No filter or cursor persistence across sessions.
- Only one color theme (GitHub Dark).

[Unreleased]: https://github.com/cwklurks/dockpose/compare/v0.1.6...HEAD
[0.1.6]: https://github.com/cwklurks/dockpose/releases/tag/v0.1.6
[0.1.0]: https://github.com/cwklurks/dockpose/releases/tag/v0.1.0
