package docker

import (
	"testing"
	"time"

	"github.com/docker/docker/api/types"
)

// Docker's ContainerList returns a human-readable Status string like
// "Up 5 minutes (healthy)" or "Up 10 seconds (unhealthy)". We parse it
// to avoid an O(N) ContainerInspect fan-out on every poll tick.
func TestParseHealthFromStatus(t *testing.T) {
	cases := []struct {
		name     string
		status   string
		state    string
		wantStat string
		wantHlth string
	}{
		{"healthy", "Up 5 minutes (healthy)", "running", "running", "healthy"},
		{"unhealthy", "Up 10 minutes (unhealthy)", "running", "unhealthy", "unhealthy"},
		{"starting", "Up 3 seconds (health: starting)", "running", "starting", "starting"},
		{"running no healthcheck", "Up 2 hours", "running", "running", ""},
		{"exited", "Exited (0) 2 minutes ago", "exited", "stopped", ""},
		{"created", "Created", "created", "stopped", ""},
		{"paused", "Up 4 minutes (Paused)", "paused", "paused", ""},
		{"running with parens not health", "Up 1 hour (strange thing)", "running", "running", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotStat, gotHlth := classifyContainer(tc.state, tc.status)
			if gotStat != tc.wantStat {
				t.Errorf("status: got %q, want %q", gotStat, tc.wantStat)
			}
			if gotHlth != tc.wantHlth {
				t.Errorf("health: got %q, want %q", gotHlth, tc.wantHlth)
			}
		})
	}
}

func TestListContainersUsesClassification(t *testing.T) {
	// Build synthetic containers the way Docker's SDK hands them back.
	summaries := []types.Container{
		{
			ID:     "a",
			Names:  []string{"/proj-api-1"},
			Image:  "api:1",
			State:  "running",
			Status: "Up 12 minutes (unhealthy)",
			Labels: map[string]string{
				"com.docker.compose.project": "proj",
				"com.docker.compose.service": "api",
			},
			Created: time.Now().Add(-12 * time.Minute).Unix(),
		},
		{
			ID:     "b",
			Names:  []string{"/proj-db-1"},
			Image:  "postgres:16",
			State:  "running",
			Status: "Up 1 hour (healthy)",
			Labels: map[string]string{
				"com.docker.compose.project": "proj",
				"com.docker.compose.service": "db",
			},
			Created: time.Now().Add(-1 * time.Hour).Unix(),
		},
	}

	states := toStates(summaries, "proj")

	if len(states) != 2 {
		t.Fatalf("expected 2 states, got %d", len(states))
	}
	byService := map[string]ContainerState{}
	for _, s := range states {
		byService[s.Service] = s
	}
	if byService["api"].Status != "unhealthy" || byService["api"].Health != "unhealthy" {
		t.Errorf("api: got status=%q health=%q; want unhealthy/unhealthy",
			byService["api"].Status, byService["api"].Health)
	}
	if byService["db"].Status != "running" || byService["db"].Health != "healthy" {
		t.Errorf("db: got status=%q health=%q; want running/healthy",
			byService["db"].Status, byService["db"].Health)
	}
}
