package docker

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	dockerclient "github.com/docker/docker/client"
)

// New returns a Docker SDK client configured from DOCKER_HOST or current context.
func New() (*dockerclient.Client, error) {
	return dockerclient.NewClientWithOpts(
		dockerclient.FromEnv,
		dockerclient.WithAPIVersionNegotiation(),
	)
}

// NewForContext returns a Docker SDK client configured for a named Docker CLI
// context. The SDK does not read Docker's current-context setting, so dockpose
// resolves the context endpoint through the Docker CLI and configures the SDK
// explicitly.
func NewForContext(name string) (*dockerclient.Client, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return New()
	}
	host, err := contextHost(name)
	if err != nil {
		return nil, err
	}
	if host == "" {
		return New()
	}
	return dockerclient.NewClientWithOpts(
		dockerclient.WithHost(host),
		dockerclient.WithAPIVersionNegotiation(),
	)
}

func contextHost(name string) (string, error) {
	cmd := exec.Command("docker", "context", "inspect", name, "--format", "{{json .Endpoints.docker}}")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("inspect docker context %q: %w", name, err)
	}
	var endpoint struct {
		Host string `json:"Host"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(out))), &endpoint); err != nil {
		return "", fmt.Errorf("parse docker context %q: %w", name, err)
	}
	return endpoint.Host, nil
}
