package archiving

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

type confirmationResult int

const (
	confirmationNone confirmationResult = iota
	confirmationConfirmed
	confirmationCancelled
)

type confirmationModel struct {
	request app.ConfirmationRequest
	list    list.Model
	help    help.Model
	keys    keyMap
	result  confirmationResult
	width   int
	height  int
}

const listItemIndent = "  "

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
			key.WithHelp("y/enter", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("n", "q", "esc", "ctrl+c"),
			key.WithHelp("n/q/esc", "cancel"),
		),
	}
}

func newConfirmationModel(request app.ConfirmationRequest) confirmationModel {
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

	return confirmationModel{
		request: request,
		list:    l,
		help:    help.New(),
		keys:    defaultKeyMap(),
	}
}

func (m confirmationModel) Init() tea.Cmd {
	return nil
}

func (m confirmationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			m.result = confirmationConfirmed
			return m, nil
		case key.Matches(msg, m.keys.Cancel) || isCtrlC(msg):
			m.result = confirmationCancelled
			return m, nil
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

func (m confirmationModel) View() tea.View {
	var parts []string
	parts = append(parts, headerView(m.request.HeaderTitle))
	parts = append(parts, "")
	parts = append(parts, summaryView(m.request))
	parts = append(parts, "")
	parts = append(parts, m.list.View())
	parts = append(parts, m.help.View(m.keys))
	return tea.NewView(strings.Join(parts, "\n"))
}

func (m confirmationModel) Result() confirmationResult {
	return m.result
}

func (m confirmationModel) ListHeight() int {
	return m.list.Height()
}

func listHeight(windowHeight int) int {
	const reservedRows = 5
	height := windowHeight - reservedRows
	if height < 3 {
		return 3
	}
	return height
}

var headerStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("170")).Padding(0, 1).
	Foreground(lipgloss.Color("255"))

var listItemStyle = lipgloss.NewStyle()

func headerView(title string) string {
	return headerStyle.Render(title)
}

func summaryView(request app.ConfirmationRequest) string {
	summary := fmt.Sprintf("%s will be moved to %s. Press y/enter to confirm.", core.FormatSize(request.ScanResult.MoveSize), request.CompactArchiveBucket)
	if skipped := len(request.ScanResult.SkippedItems); skipped > 0 {
		summary = fmt.Sprintf("%s Skipped items: %d.", summary, skipped)
	}
	return summary
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
	_, _ = fmt.Fprint(w, listItemStyle.Render(listItemIndent+item.FilterValue()))
}
