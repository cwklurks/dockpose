// Package demo provides fixture stacks and a synthetic docker.Source so
// dockpose can be demoed without a running Docker daemon.
package demo

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/cwklurks/dockpose/internal/docker"
	"github.com/cwklurks/dockpose/internal/stack"
)

// Source is a synthetic docker.Source and stack registry.
//
// It rotates container statuses on every tick so the TUI feels alive
// without touching a real Docker daemon.
type Source struct {
	mu      sync.Mutex
	stacks  []stack.Stack
	state   map[string][]docker.ContainerState
	started time.Time
	ticks   int
}

// New returns a Source pre-populated with a realistic homelab fixture.
func New() *Source {
	s := &Source{
		state:   map[string][]docker.ContainerState{},
		started: time.Now(),
	}
	s.stacks = fixtureStacks()
	for _, st := range s.stacks {
		s.state[st.Name] = initialState(st)
	}
	return s
}

// Stacks returns the fixture stacks as parsed Stack values.
func (s *Source) Stacks() []stack.Stack {
	out := make([]stack.Stack, len(s.stacks))
	copy(out, s.stacks)
	return out
}

// Tick advances the simulation one step. Typically called from the
// app's polling loop so statuses jitter over time.
func (s *Source) Tick() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ticks++
	for name, cs := range s.state {
		for i := range cs {
			cs[i] = advance(cs[i], s.ticks, name, i)
		}
		s.state[name] = cs
	}
}

// ListContainers returns the current synthetic container state for
// projectName. Satisfies docker.Source.
func (s *Source) ListContainers(_ context.Context, projectName string) ([]docker.ContainerState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cs, ok := s.state[projectName]
	if !ok {
		return nil, nil
	}
	out := make([]docker.ContainerState, len(cs))
	copy(out, cs)
	for i := range out {
		if out[i].Status == "running" {
			out[i].Uptime = time.Since(s.started) + time.Duration(i)*13*time.Minute
		}
	}
	return out, nil
}

// fixtureStacks mirrors the spec's sample layout: media, monitoring,
// traefik, dev-api, authentik.
func fixtureStacks() []stack.Stack {
	return []stack.Stack{
		{
			Name:     "media",
			Path:     "~/homelab/media/compose.yml",
			Profiles: []string{"default", "downloads"},
			Services: []stack.Service{
				{Name: "jellyfin", Image: "jellyfin/jellyfin:10.9"},
				{Name: "sonarr", Image: "linuxserver/sonarr:latest", DependsOn: []string{"prowlarr"}},
				{Name: "radarr", Image: "linuxserver/radarr:latest", DependsOn: []string{"prowlarr"}},
				{Name: "prowlarr", Image: "linuxserver/prowlarr:latest"},
				{Name: "bazarr", Image: "linuxserver/bazarr:latest", DependsOn: []string{"sonarr", "radarr"}},
			},
		},
		{
			Name:     "monitoring",
			Path:     "~/homelab/monitoring/compose.yml",
			Profiles: []string{"default"},
			Services: []stack.Service{
				{Name: "prometheus", Image: "prom/prometheus:v2.54"},
				{Name: "grafana", Image: "grafana/grafana:11.2", DependsOn: []string{"prometheus"}},
				{Name: "node-exporter", Image: "prom/node-exporter:latest"},
				{Name: "loki", Image: "grafana/loki:3.1"},
			},
		},
		{
			Name:     "traefik",
			Path:     "~/homelab/traefik/compose.yml",
			Profiles: []string{"default"},
			Services: []stack.Service{
				{Name: "traefik", Image: "traefik:v3.1"},
			},
		},
		{
			Name:     "dev-api",
			Path:     "~/projects/dev-api/compose.yml",
			Profiles: []string{"default", "debug"},
			Services: []stack.Service{
				{Name: "api", Image: "ghcr.io/acme/api:dev", DependsOn: []string{"postgres", "redis"}},
				{Name: "worker", Image: "ghcr.io/acme/worker:dev", DependsOn: []string{"postgres", "redis"}},
				{Name: "scheduler", Image: "ghcr.io/acme/scheduler:dev", DependsOn: []string{"postgres"}},
				{Name: "postgres", Image: "postgres:16"},
				{Name: "redis", Image: "redis:7"},
				{Name: "mailhog", Image: "mailhog/mailhog:latest"},
			},
		},
		{
			Name:     "authentik",
			Path:     "~/homelab/authentik/compose.yml",
			Profiles: []string{"default", "full"},
			Services: []stack.Service{
				{Name: "server", Image: "ghcr.io/goauthentik/server:2024.8", DependsOn: []string{"postgres", "redis"}},
				{Name: "worker", Image: "ghcr.io/goauthentik/server:2024.8", DependsOn: []string{"postgres", "redis"}},
				{Name: "postgres", Image: "postgres:16"},
				{Name: "redis", Image: "redis:7"},
			},
		},
	}
}

// initialState seeds per-service container state. Each fixture stack has
// a different starting posture to show off the status indicators.
func initialState(st stack.Stack) []docker.ContainerState {
	posture := map[string]string{
		"media":      "healthy",
		"monitoring": "degraded",
		"traefik":    "healthy",
		"dev-api":    "stopped",
		"authentik":  "healthy",
	}[st.Name]

	cs := make([]docker.ContainerState, 0, len(st.Services))
	for i, svc := range st.Services {
		state := docker.ContainerState{
			ID:      fmt.Sprintf("%s_%s_demo", st.Name, svc.Name),
			Name:    fmt.Sprintf("%s-%s-1", st.Name, svc.Name),
			Image:   svc.Image,
			Status:  "running",
			Health:  "healthy",
			Uptime:  time.Duration(i+1) * 37 * time.Minute,
			Ports:   demoPortsFor(svc.Name),
			Service: svc.Name,
			Project: st.Name,
		}
		switch posture {
		case "stopped":
			state.Status = "stopped"
			state.Health = ""
			state.Uptime = 0
		case "degraded":
			if i == 1 {
				state.Status = "unhealthy"
				state.Health = "unhealthy"
			}
			if i == 3 {
				state.Status = "stopped"
				state.Health = ""
				state.Uptime = 0
			}
		}
		cs = append(cs, state)
	}
	// Deterministic order.
	sort.Slice(cs, func(i, j int) bool { return cs[i].Service < cs[j].Service })
	return cs
}

// advance applies a small state change so the demo feels live.
//
// Two motions happen on each tick:
//   - dev-api starts stopped, transitions through "starting", then
//     lands on running/healthy (the happy-path upgrade scenario).
//   - monitoring keeps one unhealthy service and one stopped service
//     so the yellow ◐ and red ○ dots remain visible throughout. They
//     aren't noise — they're what the product is for.
func advance(c docker.ContainerState, tick int, stackName string, idx int) docker.ContainerState {
	if stackName == "dev-api" {
		switch {
		case tick >= 2 && tick < 5:
			c.Status = "starting"
			c.Health = "starting"
		case tick >= 5:
			c.Status = "running"
			c.Health = "healthy"
			c.Uptime = time.Duration(idx+1)*42*time.Second + time.Duration(tick)*time.Second
		}
	}
	return c
}

func demoPortsFor(svc string) []string {
	table := map[string][]string{
		"jellyfin":      {"8096/tcp"},
		"grafana":       {"3000/tcp"},
		"prometheus":    {"9090/tcp"},
		"loki":          {"3100/tcp"},
		"node-exporter": {"9100/tcp"},
		"traefik":       {"80/tcp", "443/tcp"},
		"api":           {"8080/tcp"},
		"postgres":      {"5432/tcp"},
		"redis":         {"6379/tcp"},
		"mailhog":       {"8025/tcp"},
		"server":        {"9000/tcp"},
	}
	if p, ok := table[svc]; ok {
		return p
	}
	return nil
}
