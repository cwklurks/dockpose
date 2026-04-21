package docker

import (
	"bufio"
	"context"
	"fmt"
	"strconv"

	"github.com/docker/docker/api/types/container"
)

// StreamLogs streams container log lines on the returned channel. The channel
// closes when the context is cancelled, the container disconnects, or an
// unrecoverable error occurs. If tail <= 0, all history is read.
func StreamLogs(ctx context.Context, containerID string, follow bool, tail int) (<-chan string, error) {
	cli, err := New()
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}
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
		_ = cli.Close()
		return nil, fmt.Errorf("container logs: %w", err)
	}

	ch := make(chan string, 64)
	go func() {
		defer close(ch)
		defer func() { _ = rc.Close(); _ = cli.Close() }()
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
