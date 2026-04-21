package stack

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/cwklurks/dockpose/internal/docker"
)

// osStdin/osStdout/osStderr are indirections so tests can swap stdio if needed.
func osStdin() *os.File  { return os.Stdin }
func osStdout() *os.File { return os.Stdout }
func osStderr() *os.File { return os.Stderr }

func jsonMarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

// Up runs `docker compose up -d` for the compose file at path, optionally
// activating the given profiles.
func Up(ctx context.Context, path string, profiles []string) error {
	args := []string{"compose", "-f", path}
	for _, p := range profiles {
		args = append(args, "--profile", p)
	}
	args = append(args, "up", "-d")
	return runCompose(ctx, path, args)
}

// Down runs `docker compose down` for the compose file at path.
func Down(ctx context.Context, path string) error {
	return runCompose(ctx, path, []string{"compose", "-f", path, "down"})
}

// Restart runs `docker compose restart [service]`. If service is empty, all
// services are restarted.
func Restart(ctx context.Context, path, service string) error {
	args := []string{"compose", "-f", path, "restart"}
	if service != "" {
		args = append(args, service)
	}
	return runCompose(ctx, path, args)
}

// Pull runs `docker compose pull` for the compose file at path.
func Pull(ctx context.Context, path string) error {
	return runCompose(ctx, path, []string{"compose", "-f", path, "pull"})
}

// Logs streams container logs for a service in the stack via the Docker SDK.
// Returns a channel of log lines. If service is empty, logs for all services
// in the stack are multiplexed.
func Logs(ctx context.Context, path, service string) (<-chan string, error) {
	projectName := filepath.Base(filepath.Dir(path))
	cli, err := docker.New()
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}
	containers, err := docker.ListContainers(ctx, cli, projectName)
	if err != nil {
		_ = cli.Close()
		return nil, err
	}
	out := make(chan string, 128)
	var wg sync.WaitGroup
	for _, c := range containers {
		if service != "" && c.Service != service {
			continue
		}
		ch, err := docker.StreamLogs(ctx, c.ID, true, 100)
		if err != nil {
			continue
		}
		name := c.Service
		if name == "" {
			name = c.Name
		}
		wg.Add(1)
		go func(prefix string, in <-chan string) {
			defer wg.Done()
			for line := range in {
				select {
				case <-ctx.Done():
					return
				case out <- fmt.Sprintf("%s | %s", prefix, line):
				}
			}
		}(name, ch)
	}
	go func() {
		defer func() { _ = cli.Close() }()
		wg.Wait()
		close(out)
	}()
	return out, nil
}

// Stop runs `docker compose stop [service]`. If service is empty, all services are stopped.
func Stop(ctx context.Context, path, service string) error {
	args := []string{"compose", "-f", path, "stop"}
	if service != "" {
		args = append(args, service)
	}
	return runCompose(ctx, path, args)
}

// RestartService is an alias for Restart scoped to a specific service.
func RestartService(ctx context.Context, path, service string) error {
	return Restart(ctx, path, service)
}

// Shell opens an interactive /bin/sh (falling back as needed) in the container.
// It replaces the current process's stdio with the exec session, so the caller
// should use it only when the TUI has released the terminal (e.g., via tea.ExecProcess).
func Shell(ctx context.Context, containerID string) error {
	return Exec(ctx, containerID, []string{"/bin/sh"})
}

// Exec runs a one-shot command in the container attached to the current stdio.
func Exec(ctx context.Context, containerID string, cmd []string) error {
	if len(cmd) == 0 {
		return fmt.Errorf("exec: empty command")
	}
	args := append([]string{"exec", "-it", containerID}, cmd...)
	c := exec.CommandContext(ctx, "docker", args...)
	c.Stdin = osStdin()
	c.Stdout = osStdout()
	c.Stderr = osStderr()
	if err := c.Run(); err != nil {
		return fmt.Errorf("docker exec: %w", err)
	}
	return nil
}

// Inspect returns `docker inspect <id>` output as a JSON string.
func Inspect(ctx context.Context, containerID string) (string, error) {
	cli, err := docker.New()
	if err != nil {
		return "", fmt.Errorf("docker client: %w", err)
	}
	defer func() { _ = cli.Close() }()
	info, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return "", fmt.Errorf("inspect: %w", err)
	}
	b, err := jsonMarshalIndent(info, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}
	return string(b), nil
}

func runCompose(ctx context.Context, path string, args []string) error {
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = filepath.Dir(path)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker %v: %w", args, err)
	}
	return nil
}
