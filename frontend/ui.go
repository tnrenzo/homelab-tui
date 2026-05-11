package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Tab struct {
	Title  string
	Panels []*Panel
}

type Panel struct {
	Title    string
	Expanded bool
	Content  []string
}

type model struct {
	tabs      []Tab
	activeTab int
	focus     int
	wsAddr    string
	wsPort    int
	width     int
	height    int
}

type refreshTickMsg time.Time

func newModel(wsAddr string, wsPort int) model {
	return model{
		tabs:      initTabs(),
		activeTab: 0,
		focus:     0,
		wsAddr:    wsAddr,
		wsPort:    wsPort,
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

func cpuCoresPanelContent(info SystemInfo) []string {
	n := len(info.CoreUsage)

	if n == 0 {
		return []string{"CPU Cores: waiting for data..."}
	}

	cols := 2
	if n > 12 {
		cols = 4
	} else if n > 6 {
		cols = 3
	}
	rows := (n + cols - 1) / cols

	// per-column formatting
	labelWidth := 6
	barWidth := 8
	colWidth := labelWidth + 1 + barWidth + 1 + 4

	lines := make([]string, 0, rows)
	for r := 0; r < rows; r++ {
		var parts []string
		for c := 0; c < cols; c++ {
			idx := r + c*rows
			if idx >= n {
				continue
			}
			perc := int(info.CoreUsage[idx] + 0.5)
			bar := renderBar(BarOptions{
				Percentage: &perc,
				Width:      barWidth,
				SymbolSet:  "tty_up",
				ShowValue:  false,
			})
			label := fmt.Sprintf("C%02d", idx)
			pct := fmt.Sprintf("%3d%%", perc)
			parts = append(parts, fmt.Sprintf("%-*s %s %s", labelWidth, label, bar, pct))
		}
		lines = append(lines, strings.Join(parts, "  "))
	}

	header := fmt.Sprintf("Cores (%d):", n)
	sepLen := cols*colWidth + (cols-1)*2
	if sepLen < len(header) {
		sepLen = len(header)
	}
	out := []string{header, strings.Repeat("-", sepLen)}
	out = append(out, lines...)
	return out
}

func memoryPanelContent(info SystemInfo) []string {
	if info.MemTotal == 0 {
		return []string{"Memory: waiting for data..."}
	}

	m := int(info.MemTotal)
	used := int(info.MemUsed)

	return []string{
		"Memory: " + renderBar(BarOptions{
			Max:       &m,
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

func networkPanelContent(info SystemInfo) []string {
	if len(info.Network) == 0 {
		return []string{"Network: waiting for data..."}
	}

	header := fmt.Sprintf("%-10s  %12s  %12s", "Interface", "Download", "Upload")
	content := []string{header, strings.Repeat("-", len(header))}

	for _, net := range info.Network {
		line := fmt.Sprintf("%-10s  %12s  %12s", net.Interface, formatBytes(net.RxBytes), formatBytes(net.TxBytes))
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
	for t := range m.tabs {
		setPanelContentByTitle(m.tabs[t].Panels, "System", systemPanelContent(info))
		setPanelContentByTitle(m.tabs[t].Panels, "CPU", cpuPanelContent(info))
		setPanelContentByTitle(m.tabs[t].Panels, "CPU Cores", cpuCoresPanelContent(info))
		setPanelContentByTitle(m.tabs[t].Panels, "Memory", memoryPanelContent(info))
		setPanelContentByTitle(m.tabs[t].Panels, "Disk", diskPanelContent(info))
		setPanelContentByTitle(m.tabs[t].Panels, "Processes", processPanelContent())
		setPanelContentByTitle(m.tabs[t].Panels, "Networking", networkPanelContent(info))
	}

	return m
}

func initTabs() []Tab {
	info := getLatestSystemInfo()

	overviewPanels := []*Panel{
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
	}

	resourcePanels := []*Panel{
		{
			Title:    "CPU Cores",
			Expanded: true,
			Content:  cpuCoresPanelContent(info),
		},
		{
			Title:    "Processes",
			Expanded: true,
			Content:  processPanelContent(),
		},
	}

	networkPanel := []*Panel{
		{
			Title:    "Networking",
			Expanded: true,
			Content:  networkPanelContent(info),
		},
	}

	return []Tab{
		{Title: "Overview", Panels: overviewPanels},
		{Title: "Resources", Panels: resourcePanels},
		{Title: "Networking", Panels: networkPanel},
	}
}

func (m model) Init() tea.Cmd {
	go handleWS(m.wsAddr, m.wsPort)
	return refreshTickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "1", "2", "3":
			idx, _ := strconv.Atoi(msg.String())
			if idx <= len(m.tabs) {
				m.activeTab = idx - 1
				m.focus = 0
			}

		case "tab":
			m.focus = (m.focus + 1) % len(m.tabs[m.activeTab].Panels)

		case "shift+tab":
			m.focus = (m.focus - 1 + len(m.tabs[m.activeTab].Panels)) % len(m.tabs[m.activeTab].Panels)

		case "enter":
			panels := m.tabs[m.activeTab].Panels
			if m.focus < len(panels) {
				panels[m.focus].Expanded = !panels[m.focus].Expanded
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case refreshTickMsg:
		m = refreshPanels(m)
		return m, refreshTickCmd()
	}

	return m, nil
}

func (m model) View() tea.View {
	// 1. Render Tabs
	var tabStrings []string
	for i, t := range m.tabs {
		style := lipgloss.NewStyle().Padding(0, 1)
		if i == m.activeTab {
			style = style.Foreground(lipgloss.Color("205")).Bold(true).Underline(true)
		}
		tabStrings = append(tabStrings, style.Render(fmt.Sprintf("[%d] %s", i+1, t.Title)))
	}
	header := lipgloss.JoinHorizontal(lipgloss.Top, tabStrings...) + "\n\n"

	// 2. Render Panels in a Grid (2 columns)
	activePanels := m.tabs[m.activeTab].Panels
	var rows []string
	var currentRow []string

	// Calculate panel width (rough estimate for 2 columns)
	panelWidth := (m.width / 2) - 4
	if panelWidth < 40 {
		panelWidth = 40
	}

	for i, p := range activePanels {
		isFocused := i == m.focus

		// Style the panel box
		panelStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Width(panelWidth).
			Padding(0, 1)

		if isFocused {
			panelStyle = panelStyle.BorderForeground(lipgloss.Color("205"))
		}

		renderedPanel := panelStyle.Render(renderPanel(p))
		currentRow = append(currentRow, renderedPanel)

		// Every 2 panels, start a new row
		if len(currentRow) == 2 || i == len(activePanels)-1 {
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, currentRow...))
			currentRow = []string{}
		}
	}

	body := lipgloss.JoinVertical(lipgloss.Left, rows...)

	footer := "\n\n1-2: Switch Tabs | Tab: Focus | Enter: Toggle | q: Quit"

	return tea.NewView(header + body + footer)
}

func renderPanel(p *Panel) string {
	state := "open"
	if !p.Expanded {
		state = "closed"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s [%s]\n", p.Title, state))
	if !p.Expanded {
		return strings.TrimRight(b.String(), "\n")
	}

	b.WriteString(strings.Repeat("─", 40) + "\n")

	for _, line := range p.Content {
		b.WriteString(line + "\n")
	}

	return strings.TrimRight(b.String(), "\n")
}
