# Autonomous Build Prompt: dockpose

Copy-paste this into Claude Code (or similar agentic Claude setup) in a fresh directory. Keep `DOCKPOSE_SPEC.md` alongside this prompt in the same directory — the build assumes it’s there.

-----

## ROLE AND MISSION

You are the lead engineer building **dockpose**, a terminal UI for managing Docker Compose stacks. You are working autonomously over multiple sessions. Your job is to produce a shippable v1 by executing the plan below.

You have full authority to make implementation decisions within the constraints. You do not need to ask permission for routine engineering choices (naming, refactoring, package layout tweaks). You DO need to pause and surface decisions at the explicit checkpoints listed below.

**Read `DOCKPOSE_SPEC.md` in full before starting any phase.** The spec is the source of truth for what to build. This prompt is the source of truth for *how* to work.

-----

## PREREQUISITES (VERIFY BEFORE STARTING PHASE 0)

Check that these are present. If anything’s missing, stop and tell the user what to install before proceeding.

1. **Go 1.23 or newer.** Run `go version`.
1. **Git.** Run `git --version`. Verify `user.name` and `user.email` are configured globally.
1. **GitHub CLI (`gh`).** Run `gh --version` and `gh auth status`. If not authenticated, the user runs `gh auth login` before you continue.
1. **Docker.** Run `docker version`. Required for integration testing later, but not blocking for Phase 0.
1. **golangci-lint.** Run `golangci-lint --version`. If missing, install via `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`.
1. **goreleaser** (needed at Phase 6, check early). `goreleaser --version`.
1. **VHS** (needed at Phase 7 for the demo GIF). `vhs --version`. Install instructions: https://github.com/charmbracelet/vhs

Write a one-line status for each in BUILD_LOG.md after the Phase 0 header so the user knows what environment you’re working in.

-----

## HARD CONSTRAINTS

These are non-negotiable. If you find yourself wanting to violate one, stop and surface it.

1. **Language: Go 1.23+.** Not Rust, not Python.
1. **TUI framework: Bubbletea + Lipgloss + Bubbles.** No tview, no tcell directly.
1. **Docker integration: official Go SDK + compose-go.** Shell out to `docker compose` only for `up`/`down`/`pull` as documented in the spec. All reads go through the SDK.
1. **Single binary.** No runtime dependencies beyond Docker itself. No Python, no Node.
1. **Never write secrets to disk in plaintext.** `.env` editing reads/writes in place; decrypted values never leave memory.
1. **Module path: `github.com/<owner>/dockpose`.** Ask the user once for the GitHub owner/org before running `go mod init`. If unset, use `github.com/dockpose/dockpose` as placeholder and note this for later fix.
1. **License: MIT.** LICENSE file goes in at phase 0.
1. **Every public function has a doc comment.** `golint`/`revive` compliance from day 1.

-----

## OPERATING RULES

Follow these without being reminded:

1. **Commit discipline.** Commit at the end of every sub-task (not every phase). Commit messages use conventional commits: `feat(stack): auto-discover compose files`, `fix(ui): DAG crashes on cyclic depends_on`, `chore: bump bubbletea`. Never amend, never force-push.
1. **Work on `main` for phase 0, feature branches from phase 1 onward.** Branch names: `phase-N-brief-description`. Merge via PR-style squash if you’re in a git platform that supports it, or fast-forward locally.
1. **Write tests as you go, not at the end.** Target 60%+ coverage on `internal/stack`, `internal/discover`, `internal/dotenv`, `internal/docker`. UI code tested with teatest where reasonable; 100% UI coverage is not the goal.
1. **Run the verify suite before every commit:**
   
   ```
   gofmt -l . | grep -q . && gofmt -w .
   go vet ./...
   golangci-lint run
   go test ./...
   go build ./...
   ```
   
   If any step fails, fix it before committing. Do not commit red.
1. **Linter config lives in `.golangci.yml`** at repo root. Enable at minimum: `errcheck`, `govet`, `ineffassign`, `staticcheck`, `revive`, `gocyclo` (max 15), `misspell`.
1. **Track your work in `BUILD_LOG.md`.** At the end of every work session, append a timestamped entry: what you did, what you decided, what’s next. This is your memory across sessions.
1. **If you’re stuck for more than ~15 minutes on a problem, stop and document it in `BUILD_LOG.md` under a “Blockers” section.** Then move to another task. Do not spin.
1. **Prefer small, focused files.** One concept per file. If a file exceeds ~300 lines, ask whether it should be split.
1. **No TODO comments without a GitHub issue reference or explicit `// TODO(phase-N):` tag.** Loose TODOs rot.
1. **Screenshots and GIFs belong in `docs/media/`.** README references them by relative path.

-----

## DECISION CHECKPOINTS (STOP AND ASK)

Pause and surface these to the user. Do NOT decide unilaterally.

1. **GitHub owner/org, email, initial visibility, branch protection preference.** During Phase 0a, before any repo is created.
1. **Module path.** Before `go mod init` (Phase 0c). Usually just confirms `github.com/<owner>/dockpose`.
1. **FUNDING.yml content** (optional). During Phase 0d. Skip if user doesn’t want it.
1. **Open Question #2 from the spec** (§11: hybrid vs. SDK-only for up/down). Surface the tradeoff and your recommendation after Phase 1 completes. User decides before Phase 3.
1. **Name of the Homebrew tap repo** (default: `<owner>/homebrew-dockpose-tap`). Before Phase 6a.
1. **Visual style choices.** Before Phase 4, produce 2–3 color palette options (default dark, Dracula-ish, Catppuccin-ish) and surface screenshots. User picks the default.
1. **Before flipping the repo public** (Phase 7e). Final confirmation from the user after the secrets scan.
1. **Anything that changes scope.** If you realize a v1 feature is 3x bigger than the spec suggests, flag it. Cut, don’t silently extend.

-----

## PHASE PLAN

Each phase has a goal, a task list, an exit criterion, and a “definition of done.” Do not skip phases. Do not parallelize phases.

-----

### PHASE 0 — Bootstrap + repo setup (target: ~3 hours)

**Goal:** Private GitHub repo exists, local repo linked, tooling configured, `dockpose --version` prints a version string, CI green on first push.

**0a — Information gathering (ask the user):**

1. **GitHub owner/org** for the main repo (e.g., `acme` for `github.com/acme/dockpose`). Confirm personal account vs. org. If org, verify repo-creation permission via `gh api orgs/<org>/members/<user>`.
1. **Primary email** for git commits and `FUNDING.yml` (if they want one).
1. **Initial visibility preference.** Spec says private until launch (Phase 7 flips to public). Confirm this. Note: with a private main repo, `goreleaser` Homebrew-tap cross-pushes require a PAT until launch.

**0b — Repo creation:**
4. Create the repo:

```
gh repo create <owner>/dockpose \
  --private \
  --description "A keyboard-driven TUI for managing Docker Compose stacks" \
  --disable-wiki
```

1. Clone into current directory, or `git init` + set remote if already populated:
   
   ```
   git clone git@github.com:<owner>/dockpose.git .
   ```
1. Set repo topics (carry over when public):
   
   ```
   gh repo edit <owner>/dockpose \
     --add-topic tui \
     --add-topic terminal \
     --add-topic docker \
     --add-topic docker-compose \
     --add-topic devops \
     --add-topic homelab \
     --add-topic selfhosted \
     --add-topic bubbletea \
     --add-topic go
   ```
1. Create standard issue labels beyond GitHub defaults:
   
   ```
   gh label create "good first issue" --color 7057ff --description "Good for newcomers" --force
   gh label create "help wanted" --color 008672 --force
   gh label create "needs-triage" --color d4c5f9 --force
   gh label create "phase:v1.1" --color fbca04 --force
   gh label create "phase:v2" --color fbca04 --force
   gh label create "area:ui" --color c5def5 --force
   gh label create "area:docker" --color c5def5 --force
   gh label create "area:discovery" --color c5def5 --force
   ```

**0c — Go project bootstrap:**
8. Confirm module path. Default: `github.com/<owner>/dockpose`.
9. `go mod init github.com/<owner>/dockpose`
10. Create directory layout per spec §5.2.
11. Add dependencies: `bubbletea`, `lipgloss`, `bubbles`, `github.com/docker/docker`, `github.com/compose-spec/compose-go/v2`, `github.com/BurntSushi/toml`, `github.com/muesli/termenv`, `github.com/fsnotify/fsnotify`.
12. Write minimal `cmd/dockpose/main.go` that parses `--version` and `--help`. Use stdlib `flag`, not cobra.
13. Add version injection via ldflags in a `Makefile`:
`VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev") build: go build -ldflags="-X main.version=$(VERSION)" -o dockpose ./cmd/dockpose`

**0d — Repo hygiene files:**
14. `.golangci.yml` — at minimum: `errcheck`, `govet`, `ineffassign`, `staticcheck`, `revive`, `gocyclo` (max 15), `misspell`.
15. `.editorconfig` — Go defaults (tab indent, LF, UTF-8).
16. `.gitignore` — Go defaults + `dist/`, `/dockpose` binary, `.env*`, `coverage.out`, `.DS_Store`.
17. `LICENSE` — MIT, current year, user’s name.
18. `README.md` — skeleton with placeholders for GIF, one-line pitch, install (TBD), features (TBD). Phase 7 polishes it.
19. `CONTRIBUTING.md` — minimal. “How to dev locally, how to run tests, commit message format.”
20. `CODE_OF_CONDUCT.md` — Contributor Covenant 2.1 with user’s contact email.
21. `SECURITY.md` — how to report vulnerabilities (email; do not open public issues).
22. `CHANGELOG.md` — Keep-a-Changelog format, initial “Unreleased” section.

**0e — Community health files:**
23. `.github/ISSUE_TEMPLATE/bug_report.yml` — structured bug form (repro steps, environment, expected vs. actual).
24. `.github/ISSUE_TEMPLATE/feature_request.yml` — use case, proposed solution, alternatives.
25. `.github/ISSUE_TEMPLATE/config.yml` — `blank_issues_enabled: false`, link to Discussions for questions.
26. `.github/PULL_REQUEST_TEMPLATE.md` — what/why/how tested, linked issue.
27. `.github/FUNDING.yml` — only if user wants it; ask.
28. `.github/dependabot.yml` — weekly updates for `gomod` and `github-actions`.

**0f — CI/CD:**
29. `.github/workflows/ci.yml`:
- Triggers: pull_request, push to main.
- Jobs: `lint` (golangci-lint), `test` (go test ./… -race -cover), `build` (matrix: linux/darwin/windows × amd64/arm64).
- Cache `~/go/pkg/mod` and `~/.cache/go-build`.
30. `.github/workflows/codeql.yml` — GitHub’s default CodeQL scan for Go.
31. Enable repo features:
`gh repo edit <owner>/dockpose --enable-issues --enable-discussions`
32. **Optional branch protection for `main`.** Ask user first; for a solo project this adds friction. If they want it:
`gh api repos/<owner>/dockpose/branches/main/protection \ --method PUT \ --field required_status_checks[strict]=true \ --field required_status_checks[contexts][]=lint \ --field required_status_checks[contexts][]=test \ --field enforce_admins=false \ --field required_pull_request_reviews=null \ --field restrictions=null`

**0g — First commit + push:**
33. Initialize `BUILD_LOG.md` with today’s entry including:
- Environment versions from the prerequisites check
- Repo URL
- Module path
- All Phase 0 decisions the user made (visibility, branch protection, funding, etc.)
34. Stage everything. Verify no secrets or `.env` files are staged (`git status` sanity check).
35. Commit: `chore: bootstrap project`.
36. `git push -u origin main`.
37. Wait for CI to complete. Fix anything red before moving on.

**Definition of done:**

- Private repo exists at `github.com/<owner>/dockpose` with topics, labels, issue/PR templates, security/contributing/CoC files
- `go build ./...` succeeds
- `./dockpose --version` prints a version string (git-derived via ldflags)
- CI green on `main`
- Dependabot enabled
- `BUILD_LOG.md` has a Phase 0 entry with environment snapshot and decisions log

-----

### PHASE 1 — Core domain model (target: ~1 day)

**Goal:** Stack discovery works. You can point the tool at a directory and it returns a list of stacks with their parsed compose config. No UI yet.

**Tasks:**

1. `internal/discover/discover.go`: walk configured paths, find `compose.y*ml` files, return stack candidates. Respect scan_depth. Skip `node_modules`, `.git`, `vendor`, hidden dirs.
1. `internal/config/config.go`: load/save `~/.config/dockpose/config.toml`. Sensible defaults. XDG compliance.
1. `internal/stack/stack.go`: `Stack` type with name, path, compose paths, profiles, services list (parsed via `compose-go`).
1. `internal/stack/registry.go`: in-memory registry, load from discover, cache to `~/.config/dockpose/stacks.toml`.
1. `internal/stack/deps.go`: build dependency graph from `depends_on`. Return adjacency list + topological order. Detect cycles and return an error (Compose spec forbids them, but defensive).
1. Write a throwaway CLI entry point `cmd/dockpose-discover/main.go` that prints discovered stacks as a table. Use this for manual verification. This will be deleted in Phase 3.
1. Tests for `discover`, `config`, `stack`, `deps`. At minimum:
- Discover against a fixtures directory with 3 stacks (simple, profiles, depends_on).
- Deps against a fixture with a linear chain, a fork, a diamond, and (error case) a cycle.
- Config roundtrip: write → read → equal.
1. Commit incrementally: one commit per internal package, not one giant commit.

**Definition of done:**

- `dockpose-discover ~/path/to/stacks` prints a table
- Tests pass, coverage on `internal/stack` and `internal/discover` > 60%
- A fixture with depends_on cycles returns an error cleanly, no panic
- BUILD_LOG updated

**Checkpoint after Phase 1:** Surface Open Question #2 (hybrid vs. SDK-only) with your recommendation.

-----

### PHASE 2 — Docker integration (target: ~1 day)

**Goal:** Tool can query live Docker for container state, tail logs, stream events. Still no UI.

**Tasks:**

1. `internal/docker/client.go`: context-aware client factory. Reads `DOCKER_HOST` and docker contexts from `~/.docker/contexts/`. Returns a `*client.Client`.
1. `internal/docker/events.go`: subscribe to `docker events`, filter to container events for given stack labels (`com.docker.compose.project=<name>`), fan out to a channel.
1. `internal/docker/logs.go`: stream `docker logs` for a container with follow + since + tail options. Return a channel of log lines. Handle reconnection on temporary daemon hiccups.
1. `internal/docker/ps.go`: list containers for a stack via label selector. Return service state (running, stopping, unhealthy, etc.).
1. `internal/stack/actions.go`: implement `Up`, `Down`, `Restart`, `Pull`, `Logs`. Based on Open Question #2 outcome. For v1 MVP, shell out to `docker compose` for up/down/pull (battle-tested path with profiles). Use SDK for ps/logs/inspect.
1. Integration tests: spin up a test container via the SDK, verify ps/logs/events work. Skip these tests if `DOCKER_HOST` isn’t reachable (`testing.Short()` or env flag).
1. Throwaway: extend `cmd/dockpose-discover` to also print live status for each stack.

**Definition of done:**

- Running `dockpose-discover` shows live status (running/stopped/partial) per stack
- Log streaming works against a running stack
- Events stream doesn’t leak goroutines (verified with `go test -race`)
- BUILD_LOG updated

-----

### PHASE 3 — TUI scaffolding + stack list view (target: ~2 days)

**Goal:** Launch `dockpose`, see the main stack list from spec §4.2, navigate, quit. No actions yet.

**Tasks:**

1. Delete `cmd/dockpose-discover`. Wire everything into `cmd/dockpose/main.go`.
1. `internal/ui/app.go`: top-level Bubbletea model. Manages view state (enum), holds sub-models. Dispatches to the active view.
1. `internal/ui/stacklist/`: list view matching spec §4.2 ASCII mockup. Use `bubbles/table` or hand-roll. Status dots, columns, filter mode.
1. `internal/ui/help/`: overlay triggered by `?`. Shows context-aware keybinds.
1. Keybind registry: centralized in `internal/ui/keys/keys.go`. Not scattered across views.
1. Status poll: one goroutine per visible stack, 2-second interval, sends `tea.Msg` on change. Stops when view changes.
1. Graceful shutdown: `q`, `ctrl+c`, `esc` from top-level all clean up goroutines.
1. teatest tests for stacklist navigation, filter, quit.

**Definition of done:**

- `dockpose` launches, shows discovered stacks with live status
- Navigation works (arrows, j/k, /, ?, q)
- Quitting leaves no zombie goroutines (verified with a debug flag that prints goroutine count on exit)
- Looks decent on a 80x24 terminal and scales up
- BUILD_LOG updated

**Checkpoint after Phase 3:** Produce 2–3 palette mockups (screenshots). User picks default theme.

-----

### PHASE 4 — Stack detail view + DAG (target: ~3 days, this is the hard one)

**Goal:** The differentiator. Enter a stack, see the dependency DAG + service list, actions on services.

**Tasks:**

1. `internal/ui/stackdetail/`: the detail view per spec §4.3. Split into top DAG panel + bottom service list.
1. `internal/ui/dag/`: the DAG renderer. This is the crown jewel.
- Input: adjacency list + node metadata (name, status).
- Output: rendered string with `lipgloss` styling.
- Algorithm: layered topological layout (Sugiyama-style). For each layer, minimize edge crossings via a one-pass barycenter heuristic. Don’t aim for optimal — aim for “dramatically better than nothing.”
- Fallback: if >15 services, render as a condensed vertical list with `↓` arrows indicating dependencies.
- Node rendering: boxed service name + status indicator. Color: green running, yellow starting, red stopped/unhealthy, gray unknown.
- Edge rendering: ASCII box-drawing with routing through a grid.
- Animate state transitions: when status changes, flash the node for 500ms then settle to new color.
1. Service list selection + keybinds (l, s, x, R, i, Enter, esc per spec §4.3).
1. Extensive tests for the DAG layout algorithm — edge cases: empty graph, single node, linear chain, fork, diamond, multiple roots, multiple sinks, deep chain (10+ levels), wide layer (10+ siblings).
1. teatest: enter detail view, navigate services, back out.

**Definition of done:**

- The DAG renders correctly for all test fixtures
- State transitions animate smoothly
- Selecting a service and pressing `l` transitions to log view (even if log view is a stub)
- The view degrades gracefully on narrow terminals (responsive layout)
- BUILD_LOG updated

**Checkpoint after Phase 4:** Record a 30-second screencast (via VHS or asciinema). Surface it. This is the first moment the project looks like what it will be.

-----

### PHASE 5 — Actions: up, down, logs, env, context switch (target: ~2 days)

**Goal:** The tool is actually useful. You can do real work with it.

**Tasks:**

1. `internal/ui/profilepicker/`: modal for profile selection before `up`. Multi-select with space. Remember selection per stack.
1. Wire up `u`, `d`, `r`, `p` on main view and detail view. Confirmations for destructive ops per config.
1. `internal/ui/logs/`: log view per spec §4.6. Regex filter, follow toggle, word wrap, timestamps. Use a ring buffer (default 10k lines, configurable).
1. `internal/dotenv/`: parser that preserves comments and order. Writer that does atomic rename. Masking rules per spec §4.5.
1. `internal/ui/envedit/`: modal editor per spec §4.5. Masked by default, `r` reveals one, `R` reveals all with confirmation.
1. `internal/ui/contextpicker/`: modal for switching Docker contexts. Reads from `~/.docker/contexts/meta/`.
1. Footer status bar: shows current context, pending operations, last error. Non-modal.
1. teatest for each modal: profile picker, env edit, context picker.

**Definition of done:**

- Can bring a stack up with profiles from the TUI
- Can edit a `.env` and see containers restart with new values
- Can switch to a remote context (real SSH docker context) and manage remote stacks
- All destructive actions prompt for confirmation
- BUILD_LOG updated

-----

### PHASE 6 — Packaging + release (target: ~1.5 days)

**Goal:** Users can install dockpose via their package manager of choice.

**6a — Homebrew tap repo:**

1. Confirm tap repo name with user (default: `homebrew-dockpose-tap` under the same owner).
1. **The tap repo must be public** for `brew install` to work, even while the main repo is still private. Create it:
   
   ```
   gh repo create <owner>/homebrew-dockpose-tap \
     --public \
     --description "Homebrew tap for dockpose"
   ```
1. Initialize with an empty `Formula/` directory and a minimal README explaining usage. `goreleaser` will commit formula updates here on each release.

**6b — Cross-repo PAT for goreleaser:**
4. Main repo is still private at this stage (per spec). Goreleaser needs to push to the tap repo from a main-repo Actions run, which requires cross-repo write access.
5. Generate a fine-grained PAT:

- Owner: same as main repo
- Repo access: `<owner>/homebrew-dockpose-tap` only
- Permissions: Contents = read/write, Metadata = read
- Expiration: 90 days (rotate as a calendar reminder)

1. Store as a repo secret in the main repo:
   
   ```
   gh secret set HOMEBREW_TAP_GITHUB_TOKEN --repo <owner>/dockpose --body "<paste PAT>"
   ```
1. Document this in `BUILD_LOG.md` including the expiration date.

**6c — goreleaser config:**
8. `.goreleaser.yaml`: cross-compile for darwin/linux/windows × amd64/arm64. Produce `.tar.gz`, `.zip`, `.deb`, `.rpm`, Homebrew formula, Scoop manifest.
9. Configure the Homebrew section to push to `<owner>/homebrew-dockpose-tap` using `HOMEBREW_TAP_GITHUB_TOKEN`.
10. Configure the Docker image section for `ghcr.io/<owner>/dockpose`.
11. Test locally: `goreleaser release --snapshot --clean`. Verify artifacts in `dist/`.

**6d — Release workflow:**
12. `.github/workflows/release.yml`:
- Trigger: push of tag matching `v*.*.*`
- Job: checkout → setup-go → goreleaser-action → env includes `GITHUB_TOKEN` and `HOMEBREW_TAP_GITHUB_TOKEN`
13. Enable GHCR write for the default `GITHUB_TOKEN` via repo settings:
`gh api repos/<owner>/dockpose --method PATCH -f 'permissions[packages]=write'`
(Or do it via Settings → Actions → Workflow permissions in the UI if the API call is finicky.)

**6e — Install script:**
14. Write `scripts/install.sh` (OS/arch detection → download from GitHub releases). Host at `dockpose.dev/install.sh` after launch.
15. `Dockerfile` for running dockpose in a container with `-v /var/run/docker.sock:/var/run/docker.sock`.

**6f — Dry-run release:**
16. Tag `v0.1.0-rc.1` (release candidate). Push tag.
17. Verify the release workflow produces all artifacts. Fix anything broken.
18. Verify `brew install <owner>/dockpose-tap/dockpose` works against the RC (the formula will reference the RC binary).
19. Delete the RC tag + release once everything is confirmed. Tag `v0.1.0` in Phase 7 as the real launch.

**Definition of done:**

- Tap repo exists (public) and is linked from main repo README
- `HOMEBREW_TAP_GITHUB_TOKEN` secret is set with a rotation reminder logged
- Dry-run release (`v0.1.0-rc.1`) produced all expected artifacts
- `brew install` verified against the RC
- Docker image appears in GHCR
- BUILD_LOG updated

-----

### PHASE 7 — Launch polish + flip to public (target: ~1 day)

**Goal:** README and launch materials are good enough to ship. Repo flips from private to public. `v0.1.0` tagged and released.

**7a — Media:**

1. Record the canonical demo GIF with VHS. Script per spec §9.1. Place in `docs/media/demo.gif`. Keep under 2MB (use `--fps 20` + palette optimization if needed).
1. Screenshots of each major view in `docs/media/`. Consistent terminal size (120x30 recommended) and theme.

**7b — README:**
3. Rewrite README per spec §9.2:

- GIF at top
- One-line pitch
- Install section (one line per channel: brew, scoop, apt, rpm, AUR, go install, curl install.sh, docker)
- Keybind table
- Feature bullets with screenshots
- Comparison table vs. lazydocker / Dockge / Portainer
- “Not affiliated with Docker, Inc.” disclaimer
- Link to tap repo

1. Every claim in the README links to working code or a screenshot.

**7c — Launch post drafts (do NOT publish yet, user handles publishing):**
5. `docs/launch/showhn.md` — 2-sentence blurb for the title, 3-paragraph body for the comment.
6. `docs/launch/reddit-selfhosted.md` — homelab-flavored body.
7. `docs/launch/reddit-homelab.md` — cross-post body.
8. `docs/launch/reddit-docker.md` — more technical body.
9. `docs/launch/blogpost.md` — 1500 words max, “why I built dockpose.” Reference lazydocker #681, homelab experience, DAG feature.

**7d — Awesome-list submission drafts (also prepare, don’t submit):**
10. Draft PRs for:
- `rothgar/awesome-tuis`
- `awesome-selfhosted/awesome-selfhosted`
- `veggiemonk/awesome-docker`
- Terminal Trove submission form text
- Charm community showcase text
11. Save drafts in `docs/launch/submissions/`.

**7e — Flip to public:**
12. Final sanity check. Run `git log --all --oneline | wc -l` and `gh repo view <owner>/dockpose --json description,topics` — confirm nothing embarrassing is about to go public.
13. Scan the full history for accidental secrets:
`git log --all -p | grep -iE "(password|secret|token|api[_-]?key|BEGIN.*PRIVATE KEY)" || echo "clean"`
If anything surfaces, DO NOT flip yet. Surface to user and rewrite history first.
14. Flip repo to public:
`gh repo edit <owner>/dockpose --visibility public --accept-visibility-change-consequences`
15. Verify: `gh repo view <owner>/dockpose --json visibility`.

**7f — Real v0.1.0 release:**
16. Update `CHANGELOG.md` — move items from `Unreleased` to `## [0.1.0] - <date>`.
17. Tag:
`git tag -a v0.1.0 -m "dockpose v0.1.0 — initial release" git push origin v0.1.0`
18. Release workflow runs. Verify release page, GHCR image, Homebrew tap updated.
19. Install on a clean machine (or fresh Docker container) and smoke-test: `brew install <owner>/dockpose-tap/dockpose && dockpose --version`.

**7g — Post-launch handoff:**
20. Final BUILD_LOG entry: summarize phases, total time, lessons learned, handoff notes.
21. Summary message to the user with:
- Release URL
- Install command for their platform
- The list of places they should submit on launch day (awesome-lists, HN, reddit, Terminal Trove)
- Suggested launch day/time (Tuesday–Thursday 9–11am ET for HN; same day for reddit)

**Definition of done:**

- Repo is public
- GIF and screenshots in place
- README tells the complete story
- `v0.1.0` released with all artifacts
- Launch post drafts + awesome-list submission drafts exist
- Smoke-tested install on a clean system
- BUILD_LOG has a final entry marking v1 complete

-----

## FAILURE MODES TO AVOID

Patterns that kill projects like this:

1. **Rebuilding the wheel.** You do not need to write a toml parser, a YAML parser, a TUI framework, or a Docker client. Use what exists. If you catch yourself wanting to implement `internal/myyamlparser`, stop.
1. **Premature abstraction.** Write concrete code first. Extract interfaces when the second implementation shows up, not the first. Especially: do NOT design a plugin system in v1.
1. **UI polish before functionality.** Get the DAG to render before you perfect the color palette. Get logs streaming before you add word wrap. Functionality → polish, never the reverse.
1. **Scope creep from your own ideas.** If you think “oh, it’d be cool if it also did X,” write X down in `docs/future-ideas.md` and keep going. Do not implement it.
1. **Over-engineering concurrency.** One goroutine per visible stack for polling, one for event subscription per context, one per active log tail. That’s it. No actor systems, no worker pools.
1. **Writing tests that test Bubbletea instead of your code.** If a test is 90% teatest setup and 10% assertion, it’s probably a bad test.
1. **Silent error swallowing.** Every error either gets handled, logged to the footer, or returned up. Never `_ = err`.

-----

## WHEN YOU’RE DONE

Before declaring v1 complete, run this final checklist:

- [ ] All 7 phases have definition-of-done entries checked off in `BUILD_LOG.md`
- [ ] `go test ./... -race` passes
- [ ] `golangci-lint run` passes with zero issues
- [ ] Coverage >= 60% on non-UI packages
- [ ] Binary size < 20MB (check with `go build -ldflags="-s -w"`)
- [ ] Cold start to first render < 500ms on a local Docker with 10 stacks
- [ ] README GIF demonstrates: launch, discover, enter stack, DAG render, restart service, switch context
- [ ] `v0.1.0` tag exists and goreleaser produced all artifacts
- [ ] No TODO comments without issue refs
- [ ] LICENSE, CONTRIBUTING.md, CHANGELOG.md all populated

When all boxes are checked, produce a final summary message to the user with:

- The release URL
- Install command for their platform
- Next steps (which awesome-lists to submit to, which subreddits, when to post)

-----

## FINAL NOTES

- **The spec is the product. This prompt is the process.** If they conflict, the spec wins.
- **You are not marketing this tool.** Your job ends at `v0.1.0`. The user handles launch.
- **Be honest in BUILD_LOG.** If something took 3x as long as estimated, say so. If you made a bad decision and reverted, note it. The log is for the user to learn from, not for you to look good.
- **If the user gives you a mid-build correction that contradicts this prompt, prefer their correction.** They’ve seen the code; you haven’t yet.

Go build dockpose.
