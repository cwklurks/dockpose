package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

// ComposeProjectLabel is the label key for the compose project name.
const ComposeProjectLabel = "com.docker.compose.project"

// ComposeServiceLabel is the label key for the compose service name.
const ComposeServiceLabel = "com.docker.compose.service"

// Event is a docker event for a compose-managed container.
type Event struct {
	Time      time.Time
	Type      string
	Action    string
	ActorID   string
	Service   string
	Container string
	Project   string
}

// SubscribeEvents subscribes to docker events for a compose project. The returned
// channel closes when the context is cancelled or the daemon stream errors.
func SubscribeEvents(ctx context.Context, stackName string) (<-chan Event, error) {
	cli, err := New()
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}
	f := filters.NewArgs()
	if stackName != "" {
		f.Add("label", ComposeProjectLabel+"="+stackName)
	} else {
		f.Add("label", ComposeProjectLabel)
	}
	msgs, errs := cli.Events(ctx, types.EventsOptions{Filters: f})

	out := make(chan Event, 32)
	go func() {
		defer close(out)
		defer func() { _ = cli.Close() }()
		for {
			select {
			case <-ctx.Done():
				return
			case err := <-errs:
				_ = err
				return
			case m, ok := <-msgs:
				if !ok {
					return
				}
				ev := Event{
					Time:    time.Unix(0, m.TimeNano),
					Type:    string(m.Type),
					Action:  string(m.Action),
					ActorID: m.Actor.ID,
					Service: m.Actor.Attributes[ComposeServiceLabel],
					Project: m.Actor.Attributes[ComposeProjectLabel],
				}
				select {
				case <-ctx.Done():
					return
				case out <- ev:
				}
			}
		}
	}()
	return out, nil
}
