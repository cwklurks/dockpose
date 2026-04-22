package ui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cwklurks/dockpose/internal/docker"
	"github.com/cwklurks/dockpose/internal/stack"
	"github.com/cwklurks/dockpose/internal/ui/contextpicker"
	"github.com/cwklurks/dockpose/internal/ui/dag"
	"github.com/cwklurks/dockpose/internal/ui/envedit"
	"github.com/cwklurks/dockpose/internal/ui/logview"
	"github.com/cwklurks/dockpose/internal/ui/profilepicker"
	"github.com/cwklurks/dockpose/internal/ui/stackdetail"
	"github.com/cwklurks/dockpose/internal/ui/theme"
)

// View identifies the current top-level app screen.
type View string

// View identifiers for the main TUI screens.
const (
	ViewStackList     View = "stacklist"
	ViewDetail        View = "stackdetail"
	ViewLogs          View = "logs"
	ViewInspect       View = "inspect"
	ViewProfilePicker View = "profilepicker"
	ViewEnvEditor     View = "enveditor"
	ViewContextPicker View = "contextpicker"
)

// tickMsg triggers a refresh of docker container states.
type tickMsg struct{}

// actionResultMsg is emitted after an async service action completes.
type actionResultMsg struct {
	err error
}

// demoTicker is the Source's optional Tick hook (implemented by demo.Source).
type demoTicker interface {
	Tick()
}

// AppModel is the root Bubble Tea model for the dockpose TUI.
type AppModel struct {
	CurrentView View

	stacks   []stack.Stack
	source   docker.Source
	demoMode bool

	cursor      int
	selectedIdx int
	statuses    map[string]string

	width  int
	height int

	filter        string
	filterEditing bool
	showHelp      bool

	detail        *stackdetail.ServiceDetailModel
	logModel      *logview.LogModel
	inspectText   string
	lastActionMsg string
	lastActionAt  time.Time

	profilePicker *profilepicker.Model
	envEditor     *envedit.Model
	contextPicker *contextpicker.Model
}

// NewAppModel returns the initial application state with no data source.
func NewAppModel() AppModel {
	return AppModel{
		CurrentView: ViewStackList,
		statuses:    make(map[string]string),
		width:       100,
		height:      30,
	}
}

// NewAppModelWithSource returns an AppModel wired to the given docker.Source.
// If demoMode is true, destructive actions are stubbed out.
func NewAppModelWithSource(src docker.Source, demoMode bool) AppModel {
	m := NewAppModel()
	m.source = src
	m.demoMode = demoMode
	return m
}

// Init satisfies tea.Model.
func (m AppModel) Init() tea.Cmd {
	return PollingCmd(time.Second * 2)
}

// Update satisfies tea.Model.
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tickMsg:
		return m.handleTick(msg)
	case actionResultMsg:
		return m.handleActionResult(msg)
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m AppModel) handleTick(_ tickMsg) (tea.Model, tea.Cmd) {
	if t, ok := m.source.(demoTicker); ok {
		t.Tick()
	}
	if m.source != nil {
		ctx := context.Background()
		m.refreshStackStatuses(ctx)
		if m.CurrentView == ViewDetail && m.detail != nil {
			m.detail.Refresh(ctx)
		}
	}
	return m, PollingCmd(time.Second * 2)
}

func (m AppModel) handleActionResult(msg actionResultMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.flash("error: " + msg.err.Error())
	} else {
		m.flash("ok")
	}
	if m.detail != nil {
		m.detail.Refresh(context.Background())
	}
	return m, nil
}

func (m *AppModel) flash(s string) {
	m.lastActionMsg = s
	m.lastActionAt = time.Now()
}

func (m AppModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.showHelp {
		if k := msg.String(); k == "esc" || k == "?" || k == "q" {
			m.showHelp = false
		}
		return m, nil
	}
	if m.filterEditing {
		return m.handleFilterKey(msg)
	}
	if msg.String() == "?" {
		m.showHelp = true
		return m, nil
	}
	switch m.CurrentView {
	case ViewProfilePicker:
		return m.handleProfilePicker(msg)
	case ViewEnvEditor:
		return m.handleEnvEditor(msg)
	case ViewContextPicker:
		return m.handleContextPicker(msg)
	case ViewDetail:
		return m.updateDetail(msg)
	case ViewInspect:
		return m.handleInspectKey(msg)
	case ViewLogs:
		return m.handleLogsKey(msg)
	default:
		return m.handleStackListKey(msg)
	}
}

func (m AppModel) handleInspectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if k := msg.String(); k == "esc" || k == "q" {
		m.CurrentView = ViewDetail
	}
	return m, nil
}

func (m AppModel) handleLogsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.logModel != nil {
		m.logModel.Update(msg)
	}
	if k := msg.String(); k == "esc" || k == "q" || k == "ctrl+c" {
		if m.logModel != nil {
			m.logModel.Stop()
		}
		m.CurrentView = ViewDetail
	}
	return m, nil
}

func (m AppModel) handleFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter, tea.KeyEsc:
		m.filterEditing = false
		if msg.Type == tea.KeyEsc {
			m.filter = ""
		}
		m.cursor = 0
		return m, nil
	case tea.KeyBackspace:
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
		}
	case tea.KeyRunes, tea.KeySpace:
		m.filter += msg.String()
	}
	return m, nil
}

// stackListAction maps a key to a stack-action handler. These are
// destructive or modal handlers that may emit a Cmd, so they get
// dispatched through a table to keep handleStackListKey readable.
var stackListActions = map[string]func(AppModel) (tea.Model, tea.Cmd){
	"enter": AppModel.openStackDetail,
	"u":     AppModel.handleStackUp,
	"d":     AppModel.handleStackDown,
	"r":     AppModel.handleStackRestart,
	"p":     AppModel.handleStackPull,
	"e":     AppModel.openEnvEditor,
	"c":     AppModel.openContextPicker,
}

func (m AppModel) handleStackListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()
	if k == "q" || k == "ctrl+c" {
		return m, tea.Quit
	}
	if fn, ok := stackListActions[k]; ok {
		return fn(m)
	}
	m = m.applyNavKey(k)
	return m, nil
}

// applyNavKey handles non-action keys: navigation, filter entry,
// and esc. Returns the (possibly mutated) model.
func (m AppModel) applyNavKey(k string) AppModel {
	switch k {
	case "/":
		m.filterEditing = true
		m.filter = ""
	case "j", "down":
		if m.cursor < len(m.visibleStacks())-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "g":
		m.cursor = 0
	case "G":
		if n := len(m.visibleStacks()); n > 0 {
			m.cursor = n - 1
		}
	case "esc":
		m.filter = ""
		m.cursor = 0
		m.CurrentView = ViewStackList
	}
	return m
}

func (m AppModel) selectedStack() (stack.Stack, bool) {
	visible := m.visibleStacks()
	if m.cursor < 0 || m.cursor >= len(visible) {
		return stack.Stack{}, false
	}
	return visible[m.cursor], true
}

func (m AppModel) openStackDetail() (tea.Model, tea.Cmd) {
	st, ok := m.selectedStack()
	if !ok {
		return m, nil
	}
	for i, s := range m.stacks {
		if s.Name == st.Name && s.Path == st.Path {
			m.selectedIdx = i
			break
		}
	}
	m.CurrentView = ViewDetail
	d := stackdetail.New(st, m.source)
	m.detail = &d
	return m, nil
}

func (m AppModel) handleStackUp() (tea.Model, tea.Cmd) {
	st, ok := m.selectedStack()
	if !ok {
		return m, nil
	}
	if m.demoMode {
		m.flash("demo: would `compose up` " + st.Name)
		return m, nil
	}
	if len(st.Profiles) > 1 {
		p := profilepicker.New(st.Name, st.Profiles)
		m.profilePicker = &p
		m.CurrentView = ViewProfilePicker
		return m, nil
	}
	return m, runAction(func(ctx context.Context) error {
		return stack.Up(ctx, st.Path, nil)
	})
}

func (m AppModel) handleStackDown() (tea.Model, tea.Cmd) {
	st, ok := m.selectedStack()
	if !ok {
		return m, nil
	}
	if m.demoMode {
		m.flash("demo: would `compose down` " + st.Name)
		return m, nil
	}
	return m, runAction(func(ctx context.Context) error { return stack.Down(ctx, st.Path) })
}

func (m AppModel) handleStackRestart() (tea.Model, tea.Cmd) {
	st, ok := m.selectedStack()
	if !ok {
		return m, nil
	}
	if m.demoMode {
		m.flash("demo: would `compose restart` " + st.Name)
		return m, nil
	}
	return m, runAction(func(ctx context.Context) error { return stack.Restart(ctx, st.Path, "") })
}

func (m AppModel) handleStackPull() (tea.Model, tea.Cmd) {
	st, ok := m.selectedStack()
	if !ok {
		return m, nil
	}
	if m.demoMode {
		m.flash("demo: would `compose pull` " + st.Name)
		return m, nil
	}
	return m, runAction(func(ctx context.Context) error { return stack.Pull(ctx, st.Path) })
}

func (m AppModel) openEnvEditor() (tea.Model, tea.Cmd) {
	st, ok := m.selectedStack()
	if !ok {
		return m, nil
	}
	if m.demoMode {
		m.flash("demo: .env editor disabled")
		return m, nil
	}
	e := envedit.New(st.Name, st.Path)
	m.envEditor = &e
	m.CurrentView = ViewEnvEditor
	return m, nil
}

func (m AppModel) openContextPicker() (tea.Model, tea.Cmd) {
	if m.demoMode {
		m.flash("demo: context switching disabled")
		return m, nil
	}
	ctx := contextpicker.New()
	m.contextPicker = &ctx
	m.CurrentView = ViewContextPicker
	return m, nil
}

// updateDetail handles keyboard input on the detail view.
func (m AppModel) updateDetail(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.detail == nil {
		m.CurrentView = ViewStackList
		return m, nil
	}
	switch key.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.CurrentView = ViewStackList
		return m, nil
	case "j", "down", "k", "up":
		d, _ := m.detail.Update(key)
		nd := d.(stackdetail.ServiceDetailModel)
		m.detail = &nd
		return m, nil
	case "x", "R", "s", "l", "i":
		return m.handleServiceAction(key)
	}
	return m, nil
}

func (m *AppModel) handleServiceAction(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	info, ok := m.detail.SelectedInfo()
	if !ok {
		return m, nil
	}
	if m.demoMode && key.String() != "i" && key.String() != "l" {
		m.flash("demo: action disabled (" + key.String() + ")")
		return m, nil
	}
	path := m.stacks[m.selectedIdx].Path

	switch key.String() {
	case "x":
		return m, runAction(func(ctx context.Context) error {
			return stack.Stop(ctx, path, info.ServiceName)
		})
	case "R":
		return m, runAction(func(ctx context.Context) error {
			return stack.RestartService(ctx, path, info.ServiceName)
		})
	case "s":
		if info.ContainerID == "" {
			m.flash("no container for service")
			return m, nil
		}
		c := exec.Command("docker", "exec", "-it", info.ContainerID, "/bin/sh")
		return m, tea.ExecProcess(c, func(err error) tea.Msg {
			return actionResultMsg{err: err}
		})
	case "l":
		return m, m.openLogViewer(info)
	case "i":
		return m, m.inspectContainer(info)
	}
	return m, nil
}

func (m *AppModel) openLogViewer(info stackdetail.ServiceInfo) tea.Cmd {
	if m.demoMode {
		demoLines := make(chan string, 16)
		go func() {
			defer close(demoLines)
			lines := []string{
				"[demo] streaming synthetic logs for " + info.ServiceName,
				"INFO  starting up...",
				"INFO  listening on :8080",
				"DEBUG handled GET /healthz",
				"WARN  cache miss for key=user:42",
				"INFO  reconciled state in 12ms",
			}
			for _, l := range lines {
				demoLines <- l
				time.Sleep(150 * time.Millisecond)
			}
		}()
		svcLogs := []logview.ServiceLog{{ServiceName: info.ServiceName, ContainerID: "demo", Lines: demoLines}}
		lm := logview.NewLogModel(svcLogs)
		m.logModel = &lm
		m.logModel.Start(context.Background())
		m.CurrentView = ViewLogs
		return nil
	}
	if info.ContainerID == "" {
		m.flash("no container for service")
		return nil
	}
	ctx := context.Background()
	logCh, err := docker.StreamLogs(ctx, info.ContainerID, true, 200)
	if err != nil {
		m.flash("logs: " + err.Error())
		return nil
	}
	svcLogs := []logview.ServiceLog{{
		ServiceName: info.ServiceName,
		ContainerID: info.ContainerID,
		Lines:       logCh,
	}}
	lm := logview.NewLogModel(svcLogs)
	m.logModel = &lm
	m.logModel.Start(ctx)
	m.CurrentView = ViewLogs
	return nil
}

func (m *AppModel) inspectContainer(info stackdetail.ServiceInfo) tea.Cmd {
	if m.demoMode {
		m.inspectText = fmt.Sprintf(`{
  "Id": "%s",
  "Name": "%s",
  "Image": "%s",
  "State": {
    "Status": "%s",
    "Health": { "Status": "%s" }
  },
  "Ports": %v,
  "DemoMode": true
}`, info.ContainerID, info.ServiceName, info.Image, info.Status, info.Health, info.Ports)
		m.CurrentView = ViewInspect
		return nil
	}
	if info.ContainerID == "" {
		m.flash("no container for service")
		return nil
	}
	out, err := stack.Inspect(context.Background(), info.ContainerID)
	if err != nil {
		m.flash("inspect: " + err.Error())
		return nil
	}
	m.inspectText = out
	m.CurrentView = ViewInspect
	return nil
}

// runAction runs fn in a goroutine, returning its result as actionResultMsg.
func runAction(fn func(ctx context.Context) error) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		return actionResultMsg{err: fn(ctx)}
	}
}

// PollingCmd returns a tea.Cmd that sends a tickMsg after each interval.
func PollingCmd(interval time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(interval)
		return tickMsg{}
	}
}

// refreshStackStatuses queries the data source for container status of each stack.
func (m *AppModel) refreshStackStatuses(ctx context.Context) {
	for i := range m.stacks {
		containers, err := m.source.ListContainers(ctx, m.stacks[i].Name)
		if err != nil {
			m.statuses[m.stacks[i].Name] = "error"
			continue
		}
		m.statuses[m.stacks[i].Name] = aggregateStatus(containers)
	}
}

// aggregateStatus returns a summary string from container states.
func aggregateStatus(containers []docker.ContainerState) string {
	if len(containers) == 0 {
		return "no containers"
	}
	running, stopped, unhealthy := 0, 0, 0
	for _, c := range containers {
		switch c.Status {
		case "running":
			running++
		case "stopped":
			stopped++
		default:
			unhealthy++
		}
	}
	total := len(containers)
	if unhealthy > 0 {
		return fmt.Sprintf("%d/%d degraded", running, total)
	}
	if running == 0 {
		return fmt.Sprintf("0/%d down", total)
	}
	if running < total {
		return fmt.Sprintf("%d/%d partial", running, total)
	}
	return fmt.Sprintf("%d/%d up", running, total)
}

// Refresh forces a synchronous refresh of cached statuses. Useful for
// scripted scenarios (e.g. the recorder) that don't run the polling loop.
func (m *AppModel) Refresh(ctx context.Context) {
	if m.source == nil {
		return
	}
	if t, ok := m.source.(demoTicker); ok {
		t.Tick()
	}
	m.refreshStackStatuses(ctx)
	if m.detail != nil {
		m.detail.Refresh(ctx)
	}
}

// SetStacks sets the list of discovered stacks.
func (m *AppModel) SetStacks(stacks []stack.Stack) {
	out := make([]stack.Stack, len(stacks))
	copy(out, stacks)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	m.stacks = out
}

// statusFor returns the cached status for a stack, or a default.
func (m *AppModel) statusFor(st stack.Stack) string {
	if s, ok := m.statuses[st.Name]; ok {
		return s
	}
	return "unknown"
}

// visibleStacks applies the active filter.
func (m AppModel) visibleStacks() []stack.Stack {
	if m.filter == "" {
		return m.stacks
	}
	q := strings.ToLower(m.filter)
	out := make([]stack.Stack, 0, len(m.stacks))
	for _, st := range m.stacks {
		if strings.Contains(strings.ToLower(st.Name), q) || strings.Contains(strings.ToLower(st.Path), q) {
			out = append(out, st)
		}
	}
	return out
}

// handleProfilePicker handles keyboard input for the profile picker modal.
func (m *AppModel) handleProfilePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.profilePicker == nil {
		m.CurrentView = ViewStackList
		return m, nil
	}
	p, cmd := m.profilePicker.Update(msg)
	*m.profilePicker = p.(profilepicker.Model)
	if !m.profilePicker.IsActive() {
		profiles := m.profilePicker.SelectedProfiles()
		st, _ := m.selectedStack()
		m.CurrentView = ViewStackList
		m.profilePicker = nil
		return m, runAction(func(ctx context.Context) error {
			return stack.Up(ctx, st.Path, profiles)
		})
	}
	return m, cmd
}

// handleEnvEditor handles keyboard input for the .env editor modal.
func (m *AppModel) handleEnvEditor(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.envEditor == nil {
		m.CurrentView = ViewStackList
		return m, nil
	}
	e, cmd := m.envEditor.Update(msg)
	*m.envEditor = e.(envedit.Model)
	if !m.envEditor.IsActive() {
		m.CurrentView = ViewStackList
		m.envEditor = nil
	}
	return m, cmd
}

// handleContextPicker handles keyboard input for the Docker context switcher modal.
func (m *AppModel) handleContextPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.contextPicker == nil {
		m.CurrentView = ViewStackList
		return m, nil
	}
	ctx, cmd := m.contextPicker.Update(msg)
	*m.contextPicker = ctx.(contextpicker.Model)
	if !m.contextPicker.IsActive() {
		if m.contextPicker.SelectedContext() != "" {
			m.flash("context: " + m.contextPicker.SelectedContext())
		}
		m.CurrentView = ViewStackList
		m.contextPicker = nil
	}
	return m, cmd
}

// View implements the top-level view dispatcher.
func (m AppModel) View() string {
	if m.showHelp {
		return m.viewHelp()
	}
	switch m.CurrentView {
	case ViewDetail:
		return m.viewDetail()
	case ViewInspect:
		return m.viewInspect()
	case ViewLogs:
		return m.viewLogs()
	case ViewProfilePicker:
		if m.profilePicker != nil {
			return m.profilePicker.View()
		}
	case ViewEnvEditor:
		if m.envEditor != nil {
			return m.envEditor.View()
		}
	case ViewContextPicker:
		if m.contextPicker != nil {
			return m.contextPicker.View()
		}
	default:
		return m.viewStackList()
	}
	return m.viewStackList()
}

// renderHeader produces the top bar shown on every screen.
func (m AppModel) renderHeader(subtitle string) string {
	width := m.width
	if width <= 0 {
		width = 100
	}
	title := theme.TitleStyle.Render("dockpose")
	chips := []string{}
	if m.demoMode {
		chips = append(chips, theme.DemoChipStyle.Render(" demo "))
	} else {
		chips = append(chips, theme.ContextChipStyle.Render(" local "))
	}
	if subtitle != "" {
		chips = append(chips, theme.MutedStyle.Render(subtitle))
	}
	right := theme.MutedStyle.Render("? help · q quit")
	left := lipgloss.JoinHorizontal(lipgloss.Center, append([]string{title, "  "}, chips...)...)
	gap := width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}
	bar := lipgloss.JoinHorizontal(lipgloss.Center, left, strings.Repeat(" ", gap), right)
	return theme.HeaderBarStyle.Width(width).Render(bar)
}

// renderFooter renders the keybind hint + transient toast on every screen.
func (m AppModel) renderFooter(hints string) string {
	width := m.width
	if width <= 0 {
		width = 100
	}
	toast := ""
	if m.lastActionMsg != "" && time.Since(m.lastActionAt) < 4*time.Second {
		toast = "  " + theme.ToastStyle.Render("• "+m.lastActionMsg)
	}
	if m.filterEditing {
		toast = "  " + theme.WarningInlineStyle.Render("filter: "+m.filter+"_")
	} else if m.filter != "" {
		toast = "  " + theme.MutedStyle.Render("filter: "+m.filter)
	}
	body := theme.HelpStyle.Render(hints) + toast
	return theme.FooterBarStyle.Width(width).Render(body)
}

// viewStackList renders the stack list as a clean, columnar table.
func (m AppModel) viewStackList() string {
	header := m.renderHeader(fmt.Sprintf("· %d stacks", len(m.stacks)))

	visible := m.visibleStacks()

	colStatus, colName, colSvc, colProf, colUp := 10, 22, 12, 14, 10
	headerRow := fmt.Sprintf(
		"  %-*s  %-*s  %-*s  %-*s  %-*s  %s",
		colStatus, "STATUS",
		colName, "STACK",
		colSvc, "SERVICES",
		colProf, "PROFILES",
		colUp, "UPTIME",
		"PATH",
	)

	var rows []string
	rows = append(rows, theme.TableHeaderStyle.Render(headerRow))

	if len(visible) == 0 {
		rows = append(rows, theme.MutedStyle.Render("  (no stacks match filter)"))
	}

	for i, st := range visible {
		dots := renderDots(st, m.statuses[st.Name])
		statusCell := padDisplay(dots, colStatus)
		summary := summaryFor(st, m.statuses[st.Name])
		profiles := strings.Join(st.Profiles, ",")
		if profiles == "" {
			profiles = "-"
		}
		uptime := uptimeFor(st, m.statuses[st.Name])
		pathDisplay := tildify(st.Path)

		marker := "  "
		if i == m.cursor {
			marker = theme.CursorStyle.Render("▸ ")
		}
		row := fmt.Sprintf("%s%s  %-*s  %-*s  %-*s  %-*s  %s",
			marker,
			statusCell,
			colName, truncate(st.Name, colName),
			colSvc, truncate(summary, colSvc),
			colProf, truncate(profiles, colProf),
			colUp, truncate(uptime, colUp),
			theme.MutedStyle.Render(pathDisplay),
		)
		if i == m.cursor {
			rows = append(rows, theme.SelectedRowStyle.Render(row))
		} else {
			rows = append(rows, theme.NormalStyle.Render(row))
		}
	}

	body := lipgloss.JoinVertical(lipgloss.Left, rows...)
	body = theme.PanelStyle.Render(body)

	footer := m.renderFooter("↑↓ nav · enter detail · u up · d down · r restart · p pull · e .env · / filter · c context")
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

// viewDetail renders the selected stack's detail view with DAG and table.
func (m AppModel) viewDetail() string {
	if m.selectedIdx >= len(m.stacks) {
		return "No stack selected.\n"
	}
	st := m.stacks[m.selectedIdx]
	header := m.renderHeader("· " + st.Name)

	svcMap := make(map[string]stack.ServiceConfig, len(st.Services))
	statusMap := make(map[string]string, len(st.Services))
	for _, svc := range st.Services {
		svcMap[svc.Name] = stack.ServiceConfig{DependsOn: svc.DependsOn}
		statusMap[svc.Name] = svc.Status
	}
	if m.detail != nil {
		for k, v := range m.detail.ServiceStatuses {
			statusMap[k] = v
		}
	}

	var sections []string
	sections = append(sections, theme.MutedStyle.Render(tildify(st.Path)))
	sections = append(sections, "")
	sections = append(sections, theme.HeaderStyle.Render("Dependency Graph"))
	sections = append(sections, dag.RenderDependencyGraph(svcMap, statusMap))
	sections = append(sections, "")
	sections = append(sections, theme.HeaderStyle.Render("Services"))
	if m.detail != nil {
		sections = append(sections, m.detail.View())
	}
	body := theme.PanelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, sections...))

	footer := m.renderFooter("j/k move · l logs · s shell · x stop · R restart · i inspect · esc back")
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

// viewInspect renders a raw inspect JSON text view.
func (m AppModel) viewInspect() string {
	header := m.renderHeader("· inspect")
	body := theme.PanelStyle.Render(m.inspectText)
	footer := m.renderFooter("esc/q back")
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

// viewLogs renders the streaming log viewer panel.
func (m AppModel) viewLogs() string {
	if m.logModel == nil {
		return "No log model initialized.\n"
	}
	header := m.renderHeader("· logs")
	body := m.logModel.View()
	return lipgloss.JoinVertical(lipgloss.Left, header, body)
}

// viewHelp renders the full help overlay.
func (m AppModel) viewHelp() string {
	header := m.renderHeader("· help")
	sections := []string{
		theme.HeaderStyle.Render("Stack list"),
		"  ↑/k ↓/j           navigate",
		"  g / G             jump to top / bottom",
		"  enter             open stack detail",
		"  u                 up        (compose up -d, prompts for profile)",
		"  d                 down      (compose down)",
		"  r                 restart   (compose restart)",
		"  p                 pull      (compose pull)",
		"  e                 edit .env",
		"  c                 switch Docker context",
		"  /                 filter stacks (esc clears, enter applies)",
		"  ?                 toggle this help",
		"  q                 quit",
		"",
		theme.HeaderStyle.Render("Stack detail"),
		"  l                 stream logs for selected service",
		"  s                 open shell in container",
		"  x                 stop service",
		"  R                 restart service",
		"  i                 inspect container (raw JSON)",
		"  esc               back to stack list",
		"",
		theme.HeaderStyle.Render("Logs"),
		"  f                 toggle follow",
		"  t                 toggle timestamps",
		"  w                 toggle wrap",
		"  c                 clear buffer",
		"  g / G             top / bottom",
	}
	body := theme.PanelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, sections...))
	footer := m.renderFooter("esc / ? close")
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

// renderDots returns the spec's per-service status dot strip.
func renderDots(st stack.Stack, status string) string {
	// Prefer the explicit per-service status when available; the
	// aggregate string is just a hint for fallback.
	max := len(st.Services)
	if max == 0 {
		return theme.MutedStyle.Render("·")
	}
	if max > 6 {
		max = 6
	}
	// We don't have per-service status here without a Source query;
	// approximate using the aggregate "x/y up" summary.
	running, total := parseSummary(status, len(st.Services))
	if total == 0 {
		total = len(st.Services)
	}
	if running > max {
		running = max
	}
	if total > max {
		total = max
	}
	dots := strings.Repeat(theme.RunningStyle.Render("●"), running)
	rest := total - running
	if strings.Contains(status, "degraded") && rest > 0 {
		dots += theme.UnhealthyStyle.Render("◐")
		rest--
	}
	if rest > 0 {
		dots += strings.Repeat(theme.StoppedStyle.Render("○"), rest)
	}
	if dots == "" {
		dots = theme.MutedStyle.Render("·")
	}
	return dots
}

// padDisplay pads s to displayWidth columns (using lipgloss.Width to ignore ANSI).
func padDisplay(s string, w int) string {
	pad := w - lipgloss.Width(s)
	if pad <= 0 {
		return s
	}
	return s + strings.Repeat(" ", pad)
}

// summaryFor formats services as "x/y up" or "0/y down".
func summaryFor(st stack.Stack, status string) string {
	running, total := parseSummary(status, len(st.Services))
	if total == 0 {
		total = len(st.Services)
	}
	if running == 0 {
		return fmt.Sprintf("0/%d down", total)
	}
	return fmt.Sprintf("%d/%d up", running, total)
}

func parseSummary(s string, fallbackTotal int) (int, int) {
	var running, total int
	if _, err := fmt.Sscanf(s, "%d/%d", &running, &total); err == nil {
		return running, total
	}
	return 0, fallbackTotal
}

// uptimeFor surfaces a coarse uptime estimate based on the aggregate status.
func uptimeFor(st stack.Stack, status string) string {
	if strings.Contains(status, "down") || status == "" || status == "unknown" {
		return "-"
	}
	// Without a docker query here, show a placeholder; detail view has the real number.
	return "live"
}

// tildify replaces the user's home dir with "~".
func tildify(path string) string {
	home := homeDir()
	if home != "" && strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}

func homeDir() string {
	if h, ok := os.LookupEnv("HOME"); ok {
		return h
	}
	return ""
}

// truncate trims s to n display columns, appending an ellipsis if truncated.
func truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= n {
		return s
	}
	if n == 1 {
		return "…"
	}
	// Naive byte truncation is fine here: stack names/profiles are ASCII.
	return s[:n-1] + "…"
}
