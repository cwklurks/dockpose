package docker

import (
	"context"

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
