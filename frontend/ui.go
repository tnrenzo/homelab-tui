package main

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	// Color palette
	accentColor     = lipgloss.Color("#62a480")
	borderColor     = lipgloss.Color("#62a480")
	headerBg        = lipgloss.Color("#1e1e1e")
	selectedColor   = lipgloss.Color("#ffffff")
	unselectedColor = lipgloss.Color("#808080")

	// Styles
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(selectedColor).
			Background(lipgloss.Color("#0a0a0a")).
			Padding(1, 2).
			MarginBottom(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2).
			MarginBottom(1)

	panelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor)

	expandedIndicatorStyle = lipgloss.NewStyle().
				Foreground(accentColor)

	helpStyle = lipgloss.NewStyle().
			Foreground(unselectedColor).
			Align(lipgloss.Center).
			MarginTop(1)
)

type Panel struct {
	Title    string
	Expanded bool
	Content  []string
	Selected int
}

type model struct {
	width    int
	height   int
	panels   []*Panel
	focus    int // which panel has focus
	cursor   int // cursor within focused panel
	lines    []string
	selected map[int]struct{}
}

type refreshTickMsg time.Time

func newModel() model {
	return model{
		width:    120,
		height:   30,
		panels:   initPanels(),
		focus:    0,
		cursor:   0,
		lines:    []string{},
		selected: make(map[int]struct{}),
	}
}

func memoryPanelContent(info SystemInfo) []string {
	if info.MemTotal == 0 {
		return []string{"Memory: waiting for data..."}
	}

	max := int(info.MemTotal)
	used := int(info.MemTotal - info.MemFree)

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

func diskPanelContent(info SystemInfo) []string {
	if len(info.Disks) == 0 {
		return []string{"Disk: waiting for data..."}
	}

	content := make([]string, 0, len(info.Disks))
	for _, disk := range info.Disks {
		total := int(disk.Total)
		used := int(disk.Used)
		line := fmt.Sprintf("%s %s", disk.Mountpoint, renderBar(BarOptions{
			Max:       &total,
			Current:   &used,
			Width:     20,
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
	setPanelContentByTitle(m.panels, "Memory", memoryPanelContent(info))
	setPanelContentByTitle(m.panels, "Disk", diskPanelContent(info))

	if m.focus < len(m.panels) {
		maxCursor := len(m.panels[m.focus].Content) - 1
		if maxCursor < 0 {
			m.cursor = 0
		} else if m.cursor > maxCursor {
			m.cursor = maxCursor
		}
	}

	return m
}

func initPanels() []*Panel {
	info := getLatestSystemInfo()

	return []*Panel{
		{
			Title:    "CPU",
			Expanded: true,
			Content: []string{
				"CPU 0: 25.4% [████░░░░░░░░ ]",
				"CPU 1: 42.1% [████████░░░░ ]",
				"CPU 2: 18.9% [██░░░░░░░░░░ ]",
				"CPU 3: 35.7% [███████░░░░░░]",
				"Average: 30.5%",
			},
			Selected: 0,
		},
		{
			Title:    "Memory",
			Expanded: true,
			Content:  memoryPanelContent(info),
			Selected: 0,
		},
		{
			Title:    "Disk",
			Expanded: true,
			Content:  diskPanelContent(info),
			Selected: 0,
		},
		{
			Title:    "Processes",
			Expanded: true,
			Content: []string{
				"firefox (PID: 1234) - 1.2 GB",
				"go run (PID: 5678) - 234 MB",
				"bash (PID: 9012) - 8 MB",
				"system (PID: 1) - 4 MB",
			},
			Selected: 0,
		},
	}
}

func (m model) Init() tea.Cmd {
	go handleWS()
	return refreshTickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab":
			// Switch focus between panels
			m.focus = (m.focus + 1) % len(m.panels)
			m.cursor = 0

		case "shift+tab":
			// Switch focus backwards
			m.focus = (m.focus - 1 + len(m.panels)) % len(m.panels)
			m.cursor = 0

		case "enter":
			// Toggle panel expansion
			if m.focus < len(m.panels) {
				m.panels[m.focus].Expanded = !m.panels[m.focus].Expanded
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.focus < len(m.panels) && m.cursor < len(m.panels[m.focus].Content)-1 {
				m.cursor++
			}

		case "space":
			// Toggle item selection within focused panel
			if m.focus < len(m.panels) {
				_, ok := m.selected[m.cursor]
				if ok {
					delete(m.selected, m.cursor)
				} else {
					m.selected[m.cursor] = struct{}{}
				}
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
	s := ""

	// Header
	title := "󰡴 System Overview"
	s += headerStyle.Render(title) + "\n"

	// Render panels in a grid layout
	for i, panel := range m.panels {
		if i == m.focus {
			s += renderPanel(panel, m.cursor, m.focus == i, m.width) + "\n"
		} else {
			s += renderPanel(panel, m.cursor, false, m.width) + "\n"
		}
	}

	// Footer with help text
	helpText := "Tab: Navigate | Enter: Expand/Collapse | ↑↓/jk: Scroll | Space: Select | q: Quit"
	s += helpStyle.Render(helpText)

	return tea.NewView(s)
}

func renderPanel(p *Panel, cursor int, isFocused bool, width int) string {
	indicator := "▼"
	if !p.Expanded {
		indicator = "▶"
	}

	// Panel header with expand/collapse indicator
	titleLine := fmt.Sprintf("%s %s", indicator, p.Title)
	if isFocused {
		titleLine = "▸ " + titleLine
	}
	titleRendered := panelTitleStyle.Render(titleLine)

	var borderFgColor string
	if isFocused {
		borderFgColor = "#ffffff"
	} else {
		borderFgColor = "#62a480"
	}

	style := panelStyle.Copy().BorderForeground(lipgloss.Color(borderFgColor))

	if !p.Expanded {
		return style.Render(titleRendered)
	}

	// Render content
	var content strings.Builder
	content.WriteString(titleRendered + "\n")

	for i, line := range p.Content {
		prefix := "  "
		if isFocused && i == cursor {
			prefix = "→ "
		}

		content.WriteString(prefix + line + "\n")
	}

	return style.Render(content.String())
}
