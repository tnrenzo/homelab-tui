package main

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

type Panel struct {
	Title    string
	Expanded bool
	Content  []string
}

type model struct {
	panels []*Panel
	focus  int
}

type refreshTickMsg time.Time

func newModel() model {
	return model{
		panels: initPanels(),
		focus:  0,
	}
}

func formatUptime(seconds uint64) string {
	d := time.Duration(seconds) * time.Second
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	days := hours / 24
	hours = hours % 24

	if days > 0 {
		return fmt.Sprintf("%dd %02dh %02dm", days, hours, minutes)
	}

	if hours > 0 {
		return fmt.Sprintf("%dh %02dm", hours, minutes)
	}

	return fmt.Sprintf("%dm", minutes)
}

func systemPanelContent(info SystemInfo) []string {
	if info.Hostname == "" {
		return []string{"System: waiting for data..."}
	}

	return []string{
		fmt.Sprintf("Host: %s", info.Hostname),
		fmt.Sprintf("Arch: %s", info.Arch),
		fmt.Sprintf("Kernel: %s", info.KVersion),
		fmt.Sprintf("Uptime: %s", formatUptime(info.Uptime)),
	}
}

func cpuPanelContent(info SystemInfo) []string {
	if info.CPU == 0 && info.CPUL == 0 {
		return []string{"CPU: waiting for data..."}
	}

	return []string{
		fmt.Sprintf("Physical cores: %d", info.CPU),
		fmt.Sprintf("Logical cores: %d", info.CPUL),
	}
}

func memoryPanelContent(info SystemInfo) []string {
	if info.MemTotal == 0 {
		return []string{"Memory: waiting for data..."}
	}

	max := int(info.MemTotal)
	used := int(info.MemUsed)

	return []string{
		"Memory: " + renderBar(BarOptions{
			Max:       &max,
			Current:   &used,
			Width:     20,
			SymbolSet: "braille_up",
			ShowValue: true,
		}),
	}
}

func processPanelContent() []string {
	return []string{"Process data is not available yet."}
}

func formatBytes(bytes uint64) string {
	if bytes == 0 {
		return "0 B"
	}

	units := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB"}
	value := float64(bytes)
	unitIndex := 0

	for value >= 1024 && unitIndex < len(units)-1 {
		value /= 1024
		unitIndex++
	}

	if value >= 10 || unitIndex == 0 {
		return fmt.Sprintf("%.0f %s", value, units[unitIndex])
	}

	return fmt.Sprintf("%.1f %s", value, units[unitIndex])
}

func truncateLabel(label string, max int) string {
	if max <= 0 || len(label) <= max {
		return label
	}

	if max <= 1 {
		return label[:max]
	}

	return label[:max-1] + "…"
}

func diskPanelContent(info SystemInfo) []string {
	if len(info.Disks) == 0 {
		return []string{"Disk: waiting for data..."}
	}

	maxMountWidth := len("Mount")
	for _, disk := range info.Disks {
		if len(disk.Mountpoint) > maxMountWidth {
			maxMountWidth = len(disk.Mountpoint)
		}
	}
	if maxMountWidth > 16 {
		maxMountWidth = 16
	}

	header := fmt.Sprintf("%-*s  %8s  %8s  %s", maxMountWidth, "Mount", "Used", "Total", "Usage")
	content := []string{header, strings.Repeat("-", len(header))}

	for _, disk := range info.Disks {
		total := int(disk.Total)
		used := int(disk.Used)
		mount := truncateLabel(disk.Mountpoint, maxMountWidth)
		line := fmt.Sprintf("%-*s  %8s  %8s  %s", maxMountWidth, mount, formatBytes(disk.Used), formatBytes(disk.Total), renderBar(BarOptions{
			Max:       &total,
			Current:   &used,
			Width:     12,
			SymbolSet: "tty_up",
			ShowValue: true,
		}))
		content = append(content, line)
	}

	return content
}

func refreshTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return refreshTickMsg(t)
	})
}

func setPanelContentByTitle(panels []*Panel, title string, content []string) {
	for _, panel := range panels {
		if panel.Title == title {
			panel.Content = content
			return
		}
	}
}

func refreshPanels(m model) model {
	info := getLatestSystemInfo()
	setPanelContentByTitle(m.panels, "System", systemPanelContent(info))
	setPanelContentByTitle(m.panels, "CPU", cpuPanelContent(info))
	setPanelContentByTitle(m.panels, "Memory", memoryPanelContent(info))
	setPanelContentByTitle(m.panels, "Disk", diskPanelContent(info))
	setPanelContentByTitle(m.panels, "Processes", processPanelContent())

	return m
}

func initPanels() []*Panel {
	info := getLatestSystemInfo()

	return []*Panel{
		{
			Title:    "System",
			Expanded: true,
			Content:  systemPanelContent(info),
		},
		{
			Title:    "CPU",
			Expanded: true,
			Content:  cpuPanelContent(info),
		},
		{
			Title:    "Memory",
			Expanded: true,
			Content:  memoryPanelContent(info),
		},
		{
			Title:    "Disk",
			Expanded: true,
			Content:  diskPanelContent(info),
		},
		{
			Title:    "Processes",
			Expanded: true,
			Content:  processPanelContent(),
		},
	}
}

func (m model) Init() tea.Cmd {
	go handleWS("192.168.1.11", "10001")
	return refreshTickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab":
			m.focus = (m.focus + 1) % len(m.panels)

		case "shift+tab":
			m.focus = (m.focus - 1 + len(m.panels)) % len(m.panels)

		case "enter":
			if m.focus < len(m.panels) {
				m.panels[m.focus].Expanded = !m.panels[m.focus].Expanded
			}
		}

	case refreshTickMsg:
		m = refreshPanels(m)
		return m, refreshTickCmd()
	}

	return m, nil
}

func (m model) View() tea.View {
	var s strings.Builder
	title := "System Overview"
	s.WriteString(title + "\n")
	s.WriteString(strings.Repeat("=", len(title)) + "\n\n")

	for i, panel := range m.panels {
		focused := i == m.focus
		s.WriteString(renderPanel(panel, focused))
		s.WriteString("\n\n")
	}

	if s.Len() > 0 {
		s.WriteString("\n")
	}
	s.WriteString("Tab/Shift+Tab: Focus | Enter: Expand/Collapse | q: Quit")

	return tea.NewView(s.String())
}

func renderPanel(p *Panel, isFocused bool) string {
	indicator := " "
	if isFocused {
		indicator = ">"
	}

	state := "open"
	if !p.Expanded {
		state = "closed"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s %s [%s]\n", indicator, p.Title, state))
	if !p.Expanded {
		return strings.TrimRight(b.String(), "\n")
	}

	b.WriteString("  " + strings.Repeat("─", 48) + "\n")

	for _, line := range p.Content {
		b.WriteString("  " + line + "\n")
	}

	return strings.TrimRight(b.String(), "\n")
}
