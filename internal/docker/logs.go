package docker

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"sync"

	"github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"
)

// StreamLogs streams container log lines on the returned channel. The channel
// closes when the context is cancelled, the container disconnects, or an
// unrecoverable error occurs. If tail <= 0, all history is read.
func StreamLogs(ctx context.Context, containerID string, follow bool, tail int) (<-chan string, error) {
	cli, err := New()
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}
	return streamLogsWithClient(ctx, cli, containerID, follow, tail, true)
}

// StreamLogsWithClient streams container logs using an existing client. The
// caller owns the client; it is not closed when the stream ends.
func StreamLogsWithClient(ctx context.Context, cli *dockerclient.Client, containerID string, follow bool, tail int) (<-chan string, error) {
	return streamLogsWithClient(ctx, cli, containerID, follow, tail, false)
}

// StreamLogsWithContext streams logs through the Docker CLI for a selected
// Docker context. This keeps SSH contexts working because the CLI owns that
// transport.
func StreamLogsWithContext(ctx context.Context, dockerContext, containerID string, follow bool, tail int) (<-chan string, error) {
	args := []string{"logs"}
	if follow {
		args = append(args, "-f")
	}
	if tail > 0 {
		args = append(args, "--tail", strconv.Itoa(tail))
	}
	args = append(args, containerID)
	if dockerContext != "" {
		args = append([]string{"--context", dockerContext}, args...)
	}
	cmd := exec.CommandContext(ctx, "docker", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("docker logs stdout: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("docker logs stderr: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("docker %v: %w", args, err)
	}

	ch := make(chan string, 64)
	var wg sync.WaitGroup
	scan := func(scanner *bufio.Scanner) {
		defer wg.Done()
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			case ch <- scanner.Text():
			}
		}
	}
	wg.Add(2)
	go scan(bufio.NewScanner(stdout))
	go scan(bufio.NewScanner(stderr))
	go func() {
		wg.Wait()
		_ = cmd.Wait()
		close(ch)
	}()
	return ch, nil
}

func streamLogsWithClient(ctx context.Context, cli *dockerclient.Client, containerID string, follow bool, tail int, closeClient bool) (<-chan string, error) {
	tailStr := "all"
	if tail > 0 {
		tailStr = strconv.Itoa(tail)
	}
	rc, err := cli.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Tail:       tailStr,
		Timestamps: false,
	})
	if err != nil {
		if closeClient {
			_ = cli.Close()
		}
		return nil, fmt.Errorf("container logs: %w", err)
	}

	ch := make(chan string, 64)
	go func() {
		defer close(ch)
		defer func() {
			_ = rc.Close()
			if closeClient {
				_ = cli.Close()
			}
		}()
		scanner := bufio.NewScanner(rc)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			// Strip the 8-byte multiplexed log header when present.
			if len(line) >= 8 {
				b := line[0]
				if b == 0x01 || b == 0x02 {
					line = line[8:]
				}
			}
			select {
			case <-ctx.Done():
				return
			case ch <- line:
			}
		}
	}()
	return ch, nil
}
