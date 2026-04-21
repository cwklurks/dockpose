package docker

import (
	dockerclient "github.com/docker/docker/client"
)

// New returns a Docker SDK client configured from DOCKER_HOST or current context.
func New() (*dockerclient.Client, error) {
	return dockerclient.NewClientWithOpts(
		dockerclient.FromEnv,
		dockerclient.WithAPIVersionNegotiation(),
	)
}
