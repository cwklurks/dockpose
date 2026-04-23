package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	dockerclient "github.com/docker/docker/client"
)

// Source abstracts container state retrieval so the TUI can be driven by
// either a live Docker daemon or a synthetic source (e.g. demo mode).
type Source interface {
	ListContainers(ctx context.Context, projectName string) ([]ContainerState, error)
}

// ClientSource wraps a Docker SDK client to satisfy Source.
type ClientSource struct {
	Client *dockerclient.Client
}

// NewClientSource returns a Source backed by the given Docker client.
func NewClientSource(c *dockerclient.Client) *ClientSource {
	return &ClientSource{Client: c}
}

// ListContainers returns all containers for the given Compose project.
func (s *ClientSource) ListContainers(ctx context.Context, projectName string) ([]ContainerState, error) {
	return ListContainers(ctx, s.Client, projectName)
}

// ContextSource wraps Docker CLI context selection. It is used after a user
// picks a Docker context because the CLI handles context transports such as
// ssh:// that the Docker SDK does not transparently support.
type ContextSource struct {
	Context string
}

// NewContextSource returns a Source backed by `docker --context <name>`.
func NewContextSource(name string) *ContextSource {
	return &ContextSource{Context: strings.TrimSpace(name)}
}

// ListContainers returns all containers for the given Compose project through
// the Docker CLI's context-aware `ps` command.
func (s *ContextSource) ListContainers(ctx context.Context, projectName string) ([]ContainerState, error) {
	args := []string{"ps", "--all", "--filter", "label=com.docker.compose.project=" + projectName, "--format", "{{json .}}"}
	if s.Context != "" {
		args = append([]string{"--context", s.Context}, args...)
	}
	cmd := exec.CommandContext(ctx, "docker", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("docker %v: %w", args, err)
	}
	return parseContextPS(out, projectName), nil
}

type cliContainer struct {
	ID      string `json:"ID"`
	Names   string `json:"Names"`
	Image   string `json:"Image"`
	State   string `json:"State"`
	Status  string `json:"Status"`
	Ports   string `json:"Ports"`
	Labels  string `json:"Labels"`
	Running string `json:"RunningFor"`
}

func parseContextPS(data []byte, projectName string) []ContainerState {
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	states := []ContainerState{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var c cliContainer
		if err := json.Unmarshal([]byte(line), &c); err != nil {
			continue
		}
		labels := parseLabels(c.Labels)
		status, health := classifyContainer(strings.ToLower(c.State), c.Status)
		states = append(states, ContainerState{
			ID:      c.ID,
			Name:    firstName(c.Names),
			Image:   c.Image,
			Status:  status,
			Health:  health,
			Ports:   parsePorts(c.Ports),
			Service: labels["com.docker.compose.service"],
			Project: projectName,
		})
	}
	return states
}

func parseLabels(raw string) map[string]string {
	labels := map[string]string{}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, val, ok := strings.Cut(part, "=")
		if !ok {
			labels[part] = ""
			continue
		}
		labels[strings.TrimSpace(key)] = strings.TrimSpace(val)
	}
	return labels
}

func parsePorts(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func firstName(raw string) string {
	if raw == "" {
		return ""
	}
	name, _, _ := strings.Cut(raw, ",")
	return strings.TrimSpace(name)
}
