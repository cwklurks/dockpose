// Package stackdetail renders the per-stack detail view with service list and actions.
package stackdetail

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cwklurks/dockpose/internal/docker"
	"github.com/cwklurks/dockpose/internal/stack"
	"github.com/cwklurks/dockpose/internal/ui/theme"
)

// ServiceInfo is a flattened view of a service's runtime state.
type ServiceInfo struct {
	ContainerID string
	ServiceName string
	Image       string
	Status      string
	Health      string
	Uptime      time.Duration
	Ports       []string
}

// ServiceDetailModel is the per-stack detail model.
type ServiceDetailModel struct {
	Stack           stack.Stack
	Cursor          int
	ServiceStatuses map[string]string
	Infos           []ServiceInfo
	LastError       string

	source docker.Source
}

// New returns a ServiceDetailModel populated with current container state
// fetched from the given Source.
func New(st stack.Stack, src docker.Source) ServiceDetailModel {
	m := ServiceDetailModel{
		Stack:           st,
		Cursor:          0,
		ServiceStatuses: map[string]string{},
		source:          src,
	}
	m.Refresh(context.Background())
	return m
}

// Refresh queries the configured Source for current container state.
func (m *ServiceDetailModel) Refresh(ctx context.Context) {
	if m.source == nil {
		m.LastError = "no docker source configured"
		return
	}
	containers, err := m.source.ListContainers(ctx, m.Stack.Name)
	if err != nil {
		m.LastError = err.Error()
		return
	}
	byService := map[string]docker.ContainerState{}
	for _, c := range containers {
		byService[c.Service] = c
	}
	infos := make([]ServiceInfo, 0, len(m.Stack.Services))
	for _, svc := range m.Stack.Services {
		if c, ok := byService[svc.Name]; ok {
			infos = append(infos, ServiceInfo{
				ContainerID: c.ID,
				ServiceName: svc.Name,
				Image:       svc.Image,
				Status:      c.Status,
				Health:      c.Health,
				Uptime:      c.Uptime,
				Ports:       c.Ports,
			})
			m.ServiceStatuses[svc.Name] = c.Status
		} else {
			infos = append(infos, ServiceInfo{
				ServiceName: svc.Name,
				Image:       svc.Image,
				Status:      "stopped",
			})
			m.ServiceStatuses[svc.Name] = "stopped"
		}
	}
	m.Infos = infos
}

// Init satisfies tea.Model.
func (m ServiceDetailModel) Init() tea.Cmd { return nil }

// Update satisfies tea.Model — keybindings handled by parent via SelectedInfo().
func (m ServiceDetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "j", "down":
			if m.Cursor < len(m.Infos)-1 {
				m.Cursor++
			}
		case "k", "up":
			if m.Cursor > 0 {
				m.Cursor--
			}
		}
	}
	return m, nil
}

// SelectedInfo returns the currently selected ServiceInfo.
func (m ServiceDetailModel) SelectedInfo() (ServiceInfo, bool) {
	if m.Cursor < 0 || m.Cursor >= len(m.Infos) {
		return ServiceInfo{}, false
	}
	return m.Infos[m.Cursor], true
}

// View renders the service table.
func (m ServiceDetailModel) View() string {
	var b strings.Builder
	header := fmt.Sprintf("%-3s %-18s %-28s %-10s %-10s %-10s %s",
		"", "NAME", "IMAGE", "STATUS", "HEALTH", "UPTIME", "PORTS")
	b.WriteString(theme.TableHeaderStyle.Render(header))
	b.WriteString("\n")

	for i, info := range m.Infos {
		sel := "  "
		if i == m.Cursor {
			sel = "▸ "
		}
		row := fmt.Sprintf("%s %-18s %-28s %-10s %-10s %-10s %s",
			sel,
			truncate(info.ServiceName, 18),
			truncate(info.Image, 28),
			info.Status,
			info.Health,
			formatUptime(info.Uptime),
			strings.Join(info.Ports, ","),
		)
		if i == m.Cursor {
			b.WriteString(theme.SelectedRowStyle.Render(row))
		} else {
			b.WriteString(theme.NormalStyle.Render(row))
		}
		b.WriteString("\n")
	}
	if m.LastError != "" {
		b.WriteString("\n")
		b.WriteString(theme.StoppedStyle.Render("error: " + m.LastError))
		b.WriteString("\n")
	}
	return lipgloss.NewStyle().Render(b.String())
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return s[:n]
	}
	return s[:n-1] + "…"
}

func formatUptime(d time.Duration) string {
	if d <= 0 {
		return "-"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}
