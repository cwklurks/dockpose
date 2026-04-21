// Package stack defines dockpose stack models and actions.
package stack

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/compose-spec/compose-go/v2/cli"
)

// Service is a single compose service.
type Service struct {
	Name      string
	Image     string
	Health    string
	Status    string
	DependsOn []string
}

// Stack is the root model for a Compose stack.
type Stack struct {
	Name     string
	Path     string
	Profiles []string
	Services []Service
}

// ParseCompose loads a compose.yml/yaml file and returns a Stack.
func ParseCompose(path string) (*Stack, error) {
	name := filepath.Base(filepath.Dir(path))
	opts, err := cli.NewProjectOptions(
		[]string{path},
		cli.WithName(name),
		cli.WithoutEnvironmentResolution,
	)
	if err != nil {
		return nil, fmt.Errorf("compose options: %w", err)
	}
	proj, err := opts.LoadProject(context.Background())
	if err != nil {
		return nil, fmt.Errorf("load compose: %w", err)
	}

	profileSet := map[string]struct{}{}
	services := make([]Service, 0, len(proj.Services))
	for _, svc := range proj.Services {
		var deps []string
		for depName := range svc.DependsOn {
			deps = append(deps, depName)
		}
		services = append(services, Service{Name: svc.Name, Image: svc.Image, DependsOn: deps})
		for _, p := range svc.Profiles {
			profileSet[p] = struct{}{}
		}
	}
	profiles := make([]string, 0, len(profileSet))
	for p := range profileSet {
		profiles = append(profiles, p)
	}

	return &Stack{
		Name:     proj.Name,
		Path:     path,
		Profiles: profiles,
		Services: services,
	}, nil
}
