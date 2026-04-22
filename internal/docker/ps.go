package docker

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
)

// ContainerState describes the current state of a service container.
type ContainerState struct {
	ID      string
	Name    string
	Image   string
	Status  string // "running", "stopped", "unhealthy", "starting", "paused"
	Health  string // "healthy", "unhealthy", "starting", "" (no healthcheck)
	Uptime  time.Duration
	Ports   []string
	Service string // extracted from com.docker.compose.service label
	Project string // extracted from com.docker.compose.project label
}

// ListContainers returns all containers for a given compose project name.
func ListContainers(ctx context.Context, cli *dockerclient.Client, projectName string) ([]ContainerState, error) {
	args := filters.NewArgs(
		filters.Arg("label", "com.docker.compose.project="+projectName),
	)
	containers, err := cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: args,
	})
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}
	return toStates(containers, projectName), nil
}

// healthRe extracts a health keyword from Docker's human-readable
// Status string — e.g. "Up 5 minutes (healthy)", "Up 10 seconds
// (unhealthy)", "Up 3 seconds (health: starting)". The `(Paused)`
// variant is intentionally excluded: it's handled via c.State.
var healthRe = regexp.MustCompile(`\((?:health:\s*)?(healthy|unhealthy|starting)\)`)

// classifyContainer derives (status, health) from the Docker SDK's raw
// `State` + `Status` fields. Parsing `Status` lets us surface health
// without an N+1 ContainerInspect fan-out on every poll tick, which
// matters once you're managing 50+ containers across remote daemons.
func classifyContainer(state, status string) (string, string) {
	switch state {
	case "running":
		if m := healthRe.FindStringSubmatch(strings.ToLower(status)); m != nil {
			h := m[1]
			switch h {
			case "unhealthy":
				return "unhealthy", "unhealthy"
			case "starting":
				return "starting", "starting"
			case "healthy":
				return "running", "healthy"
			}
		}
		// Running container with no healthcheck defined — optimistically
		// report "running" and leave Health empty so the UI can show a
		// neutral dot rather than a misleading green one.
		return "running", ""
	case "paused":
		return "paused", ""
	case "restarting":
		return "starting", "starting"
	default:
		// "exited", "created", "dead", "removing" — all map to stopped.
		return "stopped", ""
	}
}

// toStates converts the Docker SDK's container summaries into our
// ContainerState model.
func toStates(containers []types.Container, projectName string) []ContainerState {
	states := make([]ContainerState, 0, len(containers))
	for _, c := range containers {
		status, health := classifyContainer(c.State, c.Status)

		ports := make([]string, 0, len(c.Ports))
		for _, p := range c.Ports {
			if p.PublicPort != 0 {
				ports = append(ports, fmt.Sprintf("%d/%s", p.PublicPort, p.Type))
			}
		}

		uptime := time.Duration(0)
		if c.State == "running" {
			uptime = time.Since(time.Unix(c.Created, 0))
		}

		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		service := ""
		if c.Labels != nil {
			service = c.Labels["com.docker.compose.service"]
		}

		states = append(states, ContainerState{
			ID:      c.ID,
			Name:    name,
			Image:   c.Image,
			Status:  status,
			Health:  health,
			Uptime:  uptime,
			Ports:   ports,
			Service: service,
			Project: projectName,
		})
	}
	return states
}
