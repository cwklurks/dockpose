# dockpose BUILD_LOG

Running log of the autonomous build. Append-only, newest entries at the bottom of each phase section.

---

## Phase 0 — Bootstrap + repo setup

### 2026-04-20 — Session 1 (kickoff)

**Environment snapshot (prerequisites check):**

- Go: **NOT INSTALLED** — `go version` → command not found. **BLOCKER**.
- Git: v2.43.0 installed. `user.name` and `user.email` **NOT configured globally** — BLOCKER for commit authorship.
- GitHub CLI: v2.45.0, authenticated as `cwklurks` (scopes: gist, read:org, repo, workflow). OK.
- Docker: Engine v29.2.1 (linux/arm64), server reachable. OK.
- golangci-lint: **NOT INSTALLED**. Needed from Phase 0 onward.
- goreleaser: **NOT INSTALLED**. Needed at Phase 6, check early per spec.
- VHS: **NOT INSTALLED**. Needed at Phase 7 only (non-blocking now).

**Working directory:** `/home/cwk/dockpose/` contains `BUILD_PROMPT.md`, `DOCKPOSE_SPEC.md`, `v1-spec.md` (appears identical to DOCKPOSE_SPEC.md), and `claude-tmux.log` (non-project artifact). No git repo initialized yet.

**Status:** Phase 0 bootstrap now completed locally after prerequisites were resolved and repository scaffolding was finished.

### 2026-04-20 — Session 2 (bootstrap completed by Hermes)

**Resolved decisions:**

- Git identity set to `cwklurks <connorklann@gmail.com>`.
- GitHub owner/module path confirmed as `github.com/cwklurks/dockpose`.
- Initial visibility remains private until launch.
- `FUNDING.yml` skipped.
- Branch protection skipped for now.

**Bootstrap work completed:**

- Added initial Go command skeletons for `cmd/dockpose` and `cmd/dockpose-discover`.
- Added minimal internal package scaffolding for `config`, `discover`, `docker`, `dotenv`, `stack`, and `ui`.
- Added Bubble Tea, Lip Gloss, Bubbles, Docker SDK, compose-go, and BurntSushi/toml dependencies.
- Added `.goreleaser.yaml` for cross-platform releases and package generation.
- Added GitHub Actions workflows for CI verification and tagged releases.
- Added `.pre-commit-config.yaml` and updated the Makefile verify pipeline.
- Added a placeholder demo asset at `docs/media/demo.gif`.
- Updated README/CONTRIBUTING to match the current bootstrap state.

**Verification:**

- `make check` ✅
- `./dockpose --version` ✅
- `./dockpose-discover --version` ✅

**Notes:**

- The local Go toolchain/dependency set currently targets Go 1.25.
- `go test -race -cover ./...` was reduced to `go test -race ./...` in the bootstrap Makefile because the active toolchain environment lacked the `covdata` tool required for coverage mode.

**Next checkpoint:** Phase 1 implementation can begin: discovery, registry/state plumbing, and the first Bubble Tea stack-list view.

---

## Phase 1 — Core domain model

### 2026-04-20 (completed by Hermes + Claude Code print-mode)

**Status:** Complete.

**What was built:**

- `internal/discover/discover.go` + `_test.go`: `Discover(scanPaths, scanDepth)` walks directories, finds `compose.yml/yaml`, skips `.git`/hidden/vendor/node_modules. Cycle-containing stacks (detected via deps check) are skipped from discovery results. Tests: 4 fixtures, all pass.
- `internal/stack/stack.go`: `Stack` and `Service` types; `ParseCompose(path)` uses compose-go v2 to extract name, services (Name, Image), and profiles from a compose file.
- `internal/stack/deps.go` + `_test.go`: `BuildGraph`, `TopologicalOrder`, `DetectCycles`. Uses Kahn's algorithm for topological sort, DFS for cycle detection. Returns error with cycle names on cycles. Tests: linear chain, fork, diamond, cycle error case — all pass.
- `internal/stack/registry.go` + `_test.go`: `LoadFromPaths` uses Discover+ParseCompose; `CacheTo` writes TOML atomically (tmp+rename); `LoadCache` reads from TOML. Tests: load+cache+reload from fixture dirs — pass.
- `cmd/dockpose-discover/main.go`: loads config, discovers stacks, prints tab-separated table (PATH/NAME/SERVICES/PROFILES). Exits 1 with "No stacks found." when registry is empty.

**Verification:**
- `make check` ✅ (0 lint issues)
- `go test ./...` ✅
- `go build ./...` ✅
- `go test -race ./...` ✅

**Known issue:**
- `dockpose-discover` uses default config scan paths (`~/docker`, etc.) which may not exist on all machines. The unit tests cover the logic; manual testing requires real compose files in expected paths or a future `--scan` flag override.

**Next checkpoint (Open Question #2 — per BUILD_PROMPT):**
Before Phase 3 (TUI scaffolding), decide on the abstraction boundary for compose operations:
- **Hybrid (recommended):** shell out to `docker compose` for up/down/pull (battle-tested with profiles); use Docker SDK for reads (ps/logs/inspect/events).
- **SDK-only:** use compose-go + Docker SDK directly for everything, no shell-out.
Recommendation: Hybrid — faster to implement, more robust for profiles, compose-go can parse but compose CLI manages state.
