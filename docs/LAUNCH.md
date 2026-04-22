# dockpose — launch plan

Goal: **300+ GitHub stars and ~2,000 distinct installs within 30 days of launch.**

Non-goals: being in homebrew-core, bundling a paid tier, building a web UI, competing with Portainer feature-for-feature. The point is stack-first, keyboard-first, SSH-native.

Rough calendar: pre-launch polish (Day −3 to 0) → soft launch to target communities (Day 1) → broader push (Day 2–3) → ecosystem presence (Week 1–2) → steady drumbeat (Week 3+).

---

## Phase 0 — pre-launch polish (Day −3 to 0)

Before posting anywhere, the landing experience has to be frictionless. Someone who hits the repo from a link should go from "what is this?" → "I've got it running" in under 90 seconds.

### P0 (blockers before any post)

- [x] **Clean version history.** Delete `v0.1.1`–`v0.1.5` (all CI plumbing). Keep `v0.1.0` as history and `v0.1.6` as Latest. Releases page stays scannable.
- [x] **`brew install dockpose` works.** Primary install path, verified on macOS and Linux.
- [x] **Repo is public.** Required for Homebrew, any linkbacks, and HN to even work.
- [ ] **Add social preview / OG image.** Dragging a 1280×640 PNG into GitHub Settings → General → Social preview is a 30-second job with disproportionate impact on Twitter/Bluesky/Slack/Discord link cards. The GIF's first frame exported to PNG plus "dockpose" overlaid is fine.
- [ ] **Add a `CHANGELOG.md`** with v0.1.0 notes. First-time visitors read it.
- [ ] **Pin the tagline.** GitHub About section under repo settings — 120 chars or fewer. Current: *A keyboard-driven TUI for managing Docker Compose stacks.* Add topics: `docker`, `docker-compose`, `tui`, `homelab`, `devops`, `cli`, `golang`, `bubbletea`.

### P1 (high-leverage, cheap)

- [ ] **CHANGELOG.md** in conventional format driven by git log. Goreleaser already emits it per-release; mirror it at repo root so it's the first thing in the file tree.
- [ ] **Issue templates.** `.github/ISSUE_TEMPLATE/` currently empty. Add a `bug_report.yml` and `feature_request.yml` so drive-by bug reports come in structured.
- [ ] **Two more screenshots beyond the demo GIF:** the dependency DAG view and the help overlay. Placed side-by-side in a table in the README beneath the GIF. Aesthetic reinforcement.
- [ ] **Pre-post sanity pass.** Install via `brew install dockpose` on a clean VM (or the user's laptop). Run `dockpose --demo` and every keybind. Any jank found here costs 10× more to fix once people are watching.

### P2 (nice to have)

- [ ] Proper `ARCHITECTURE.md` (1–2 pages) explaining the `Source` interface and demo parity. Becomes the "Show HN, here's how it's built" companion link.
- [ ] `asciinema.org` upload of `demo.cast` — the cast plays smoother inline on HN/Reddit than a GIF loops.

---

## Phase 1 — soft launch (Day 1)

**Day 1 is the whole ballgame.** 80% of lifetime stars happen in the first 48 hours of a launch that works. Timing and sequencing matter.

### Timing

- **Tuesday, Wednesday, or Thursday.** Never Friday (HN dies on weekends). Never Monday (inbox backlog). Avoid weeks with major releases (OpenAI dev day, Docker conference, etc.).
- **First post: 8:00 AM ET.** Beats the HN morning front-page churn and catches European readers before lunch.

### Channels — in this order

**1. r/selfhosted (primary audience).** Title: *"dockpose — a keyboard-driven TUI for managing Docker Compose stacks"*. Body: the README's "What is this" + the GIF + the one `brew install dockpose` line. Post flair: "Release."

**2. Show HN (two hours later).** Title: `Show HN: Dockpose – A keyboard-driven TUI for Docker Compose stacks` (note Show HN rules: no brackets, no emoji, no marketing). Body: 3–4 sentences on why it exists, the tech (Bubble Tea + Docker SDK + compose-go), demo mode, plus the install line. Include the GIF via imgur link (HN strips embedded assets).

**3. r/commandline.** Smaller but high signal for TUI aesthetics. Same copy, lean into the visual: the demo GIF is the headline.

**4. Charm Bracelet Discord / community channel.** They actively amplify good Bubble Tea apps. #showcase or whatever the equivalent channel is.

**5. Bluesky / Twitter thread.** 1/ problem, 2/ GIF, 3/ install command, 4/ call-to-action (star + SSH-tap the VPS). Tag `@charmbracelet`, `@dagger_io`, `@docker`.

### Posting rules

- Reply to every comment within 2 hours during business hours. Latency kills Reddit threads.
- Never get defensive. "Fair point — noted as a GitHub issue: <link>" wins every argument.
- Answer "why not lazydocker / dockge / ctop / Portainer" calmly. Link to the README's "Why?" section, don't retype it.
- If someone finds a bug, triage it live on the thread. Visible responsiveness converts skeptics to users.

### Metrics to watch (real-time)

- Stars/hour (goal: 10+/hour for first 12 hours)
- `brew install dockpose` count via `brew analytics` (lags 24h but appears)
- Clone count (GitHub → Insights → Traffic, 24h lag)
- HN rank (top 20 for 6+ hours = a good launch)

---

## Phase 2 — broader push (Day 2–3)

If Phase 1 caught fire (HN front page, r/selfhosted top post): lean in. If it was mid: still execute Phase 2, the second wave of communities has shorter memory.

- **r/docker** — same copy, lead with "stack-first" framing, because r/docker leans ops rather than homelab.
- **r/homelab** — lead with the dependency-graph view. Homelabbers love seeing their own Traefik/Authentik/Jellyfin set up mirrored.
- **r/golang** — lead with the technical angle: Bubble Tea + the `Source` interface that gives real-mode and demo-mode parity. Code quality is the selling point.
- **lobste.rs** — quality-conscious dev crowd. Invite-only but sees decent engagement for dev tools. Tag `go`, `devops`.
- **Hacker News (regular submission, not Show HN)** — only if the Show HN thread died below the front page. A week later, re-submit as a regular link with a blog post URL.

---

## Phase 3 — ecosystem presence (Week 1–2)

These are PRs and listings that compound over time — they're how someone finds you six months from now via "best Docker TUI" search.

- [ ] **awesome-selfhosted** — PR to https://github.com/awesome-selfhosted/awesome-selfhosted. They're strict about format; follow CONTRIBUTING.
- [ ] **awesome-go** — PR to https://github.com/avelino/awesome-go. Category: Utilities or DevOps Tools.
- [ ] **awesome-tuis** — PR to https://github.com/rothgar/awesome-tuis.
- [ ] **awesome-compose-like** lists — search "awesome docker" on GitHub, PR to the live ones.
- [ ] **Charm's own showcase** — they maintain a Bubble Tea community list. PR or DM.
- [ ] **AUR + Scoop + Chocolatey** — goreleaser already supports all three via one-block-each additions to `.goreleaser.yaml`. Cut v0.2.0 with all of them.

---

## Phase 4 — steady drumbeat (Week 3+)

Launches don't retain; content and response do. Once every 1–2 weeks, post something new that pulls people back.

- **A "what I learned shipping dockpose" blog post.** Syndicate to HN + Reddit. HN loves retros.
- **A "healthcheck-aware status" release.** Feature work that moves the product.
- **A screencast walking through a real workflow** (not the demo). "Here's me restarting my Jellyfin stack after a kernel update" — 90 seconds, posted to r/selfhosted.
- **Engage where people are complaining.** Search `"lazydocker"` + "stack" on Reddit/HN. Drop a non-spammy "btw, dockpose does this" comment when truly relevant.

---

## Risks and mitigations

| Risk | Mitigation |
| --- | --- |
| HN buries Show HN post fast | Post regular submission 5–7 days later with a blog post URL instead of the repo |
| Someone finds a real bug during launch | Have `gh issue list` and `dockpose --demo` open in tmux. Fix and `brew upgrade dockpose` within the hour if possible. |
| Comparison flamewar (lazydocker, dockge, Portainer) | Pre-write a non-defensive position in the README's "Why?". Link there, don't re-argue. |
| Spiky traffic on the homebrew-tap | GitHub handles it; it's a static Formula file. No mitigation needed. |
| Audience smells "AI-generated" | README is written in a human voice; don't lead with "Built with Claude" in the post. Mention if asked. |

---

## Copy drafts

### r/selfhosted body

> I've been running ~12 Compose stacks across a NAS + VPS, and the constant `cd stack && docker compose up -d && cd ..` dance finally got the better of me. Lazydocker is container-centric, Dockge is a web UI, I wanted something keyboard-first that knew about *stacks*.
>
> dockpose auto-discovers every `compose.y*ml` on your machine, shows them in one list with live status, and turns common actions into single keystrokes. Works locally and over SSH. There's a `--demo` mode if you want to poke at it without Docker running.
>
> ```
> brew tap cwklurks/tap
> brew install dockpose
> dockpose --demo
> ```
>
> Go TUI, single static binary, MIT. Roadmap: AUR/Scoop, healthcheck-aware status, container events stream. Happy to answer questions — and if there's a workflow lazydocker covers better for you, I'm curious.
>
> GitHub: https://github.com/cwklurks/dockpose

### Show HN body

> I built dockpose because running 10+ Docker Compose stacks on a homelab or VPS means a lot of cd-ing around or opening a web UI. Lazydocker is container-centric by design, Dockge is a browser tool. I wanted a keyboard-first TUI that treats *stacks* as the unit of mental work.
>
> dockpose auto-discovers every compose file, visualizes dependencies, and maps every common action to a single keybind. It works over SSH (remote Docker contexts are first-class). Built on Bubble Tea + the Docker SDK + compose-go — so parsing matches what `docker compose` itself does.
>
> There's a `--demo` mode with a synthetic homelab so you can try it without a daemon. `brew install dockpose` on macOS/Linux after tapping cwklurks/tap.
>
> https://github.com/cwklurks/dockpose

### One-liner for Bluesky/Twitter/Discord

> just shipped **dockpose** — a keyboard-driven TUI for Docker Compose stacks. k9s vibe, stack-first (not container-first), ssh-native, bubble tea. `brew install cwklurks/tap/dockpose` → `dockpose --demo`. ⭐ if you run a homelab 👇

---

## Success targets, honest

| Metric | Good | Great | Exceptional |
| --- | --- | --- | --- |
| GitHub stars (30d) | 100 | 500 | 2,000+ |
| Show HN placement | Top 50 | Front page 6h+ | #1 for an hour |
| r/selfhosted upvotes | 100 | 500 | 1,000+ |
| `brew install dockpose` count (7d) | 50 | 500 | 3,000+ |
| GitHub issues in first week | 0–3 real bugs | ~10 (healthy engagement) | ~25 (real adoption) |
| PRs in first month | 0 | 1–2 | 5+ |

If "Good" doesn't happen within 72 hours of Phase 1: the problem is usually the launch copy or timing, not the product. Retry Phase 1 in a different week with tightened copy before moving on to Phase 2.
