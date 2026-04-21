package docker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
)

// ContainerState describes the current state of a service container.
type ContainerState struct {
	ID      string
	Name    string
	Image   string
	Status  string // "running", "stopped", "unhealthy", "starting"
	Health  string // "healthy", "unhealthy", ""
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

	states := make([]ContainerState, 0, len(containers))
	for _, c := range containers {
		status := "stopped"
		if c.State == "running" {
			status = "running"
		}
		health := ""
		if c.State == "running" {
			health = "healthy"
		}

		ports := make([]string, 0)
		for _, p := range c.Ports {
			if p.PublicPort != 0 {
				ports = append(ports, fmt.Sprintf("%d/%s", p.PublicPort, p.Type))
			}
		}

		uptime := time.Since(time.Unix(c.Created, 0))
		if c.State != "running" {
			uptime = 0
		}

		name := c.Names[0]
		name = strings.TrimPrefix(name, "/")

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
	return states, nil
}
