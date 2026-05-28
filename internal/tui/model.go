package tui

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"shed/internal/app"
	"shed/internal/core"
)

type Result int

const (
	ResultNone Result = iota
	ResultConfirm
	ResultCancel
)

type Model struct {
	request app.ConfirmationRequest
	list    list.Model
	help    help.Model
	keys    keyMap
	result  Result
	width   int
	height  int
}

type keyMap struct {
	Confirm key.Binding
	Cancel  key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Confirm, k.Cancel}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

func defaultKeyMap() keyMap {
	return keyMap{
		Confirm: key.NewBinding(
			key.WithKeys("y", "enter"),
			key.WithHelp("y/enter", "archive"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("n", "q", "esc", "ctrl+c"),
			key.WithHelp("n/q/esc", "cancel"),
		),
	}
}

func NewModel(request app.ConfirmationRequest) Model {
	items := make([]list.Item, len(request.ScanResult.StaleItems))
	for i, item := range request.ScanResult.StaleItems {
		items[i] = staleListItem(item.DisplayName)
	}

	l := list.New(items, displayNameDelegate{}, 80, 10)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowFilter(false)
	l.SetShowHelp(false)
	l.SetStatusBarItemName("item", "items")

	return Model{
		request: request,
		list:    l,
		help:    help.New(),
		keys:    defaultKeyMap(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.SetWidth(msg.Width)
		m.list.SetSize(msg.Width, listHeight(msg.Height))
		return m, nil
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Confirm):
			m.result = ResultConfirm
			return m, tea.Quit
		case key.Matches(msg, m.keys.Cancel) || isCtrlC(msg):
			m.result = ResultCancel
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func isCtrlC(msg tea.KeyPressMsg) bool {
	pressed := msg.Key()
	return pressed.Mod&tea.ModCtrl != 0 && pressed.Code == 'c'
}

func (m Model) View() tea.View {
	var parts []string
	parts = append(parts, headerView(m.request.HeaderTitle))
	parts = append(parts, summaryView(m.request))
	parts = append(parts, m.list.View())
	parts = append(parts, m.help.View(m.keys))
	return tea.NewView(strings.Join(parts, "\n"))
}

func (m Model) Result() Result {
	return m.result
}

func (m Model) ListHeight() int {
	return m.list.Height()
}

func listHeight(windowHeight int) int {
	const reservedRows = 7
	height := windowHeight - reservedRows
	if height < 3 {
		return 3
	}
	return height
}

var headerStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("170")).Padding(0, 1).
	Foreground(lipgloss.Color("255"))

func headerView(title string) string {
	return headerStyle.Render(title)
}

func summaryView(request app.ConfirmationRequest) string {
	lines := []string{
		fmt.Sprintf("Move size: %s", core.FormatSize(request.ScanResult.MoveSize)),
		fmt.Sprintf("Archive bucket: %s", request.CompactArchiveBucket),
	}
	if skipped := len(request.ScanResult.SkippedItems); skipped > 0 {
		lines = append(lines, fmt.Sprintf("Skipped items: %d", skipped))
	}
	return strings.Join(lines, "\n")
}

type staleListItem string

func (item staleListItem) FilterValue() string {
	return string(item)
}

type displayNameDelegate struct{}

func (displayNameDelegate) Height() int {
	return 1
}

func (displayNameDelegate) Spacing() int {
	return 0
}

func (displayNameDelegate) Update(tea.Msg, *list.Model) tea.Cmd {
	return nil
}

func (displayNameDelegate) Render(w io.Writer, _ list.Model, _ int, item list.Item) {
	fmt.Fprint(w, item.FilterValue())
}
