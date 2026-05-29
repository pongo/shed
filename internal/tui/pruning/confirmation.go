package pruning

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"shed/internal/core"
)

type confirmationResult int

const (
	confirmationNone confirmationResult = iota
	confirmationConfirmed
	confirmationSkipped
	confirmationQuit
)

type keyMap struct {
	Confirm key.Binding
	Skip    key.Binding
	Quit    key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Confirm, k.Skip, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

type confirmationModel struct {
	scan   core.PruneScanResult
	list   list.Model
	help   help.Model
	keys   keyMap
	result confirmationResult
}

func newConfirmationModel(scan core.PruneScanResult) confirmationModel {
	items := make([]list.Item, len(scan.Candidates))
	for i, candidate := range scan.Candidates {
		items[i] = shedMonthItem(candidate.Month.Path)
	}

	l := list.New(items, shedMonthDelegate{}, 80, 10)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowFilter(false)
	l.SetShowHelp(false)

	return confirmationModel{
		scan: scan,
		list: l,
		help: help.New(),
		keys: keyMap{
			Confirm: key.NewBinding(key.WithKeys("y", "enter"), key.WithHelp("y/enter", "confirm")),
			Skip:    key.NewBinding(key.WithKeys("n", "esc"), key.WithHelp("n/esc", "skip")),
			Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q/ctrl+c", "quit")),
		},
	}
}

func (m confirmationModel) Init() tea.Cmd {
	return nil
}

func (m confirmationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.help.SetWidth(msg.Width)
		m.list.SetSize(msg.Width, listHeight(msg.Height))
		return m, nil
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Confirm):
			m.result = confirmationConfirmed
			return m, nil
		case key.Matches(msg, m.keys.Skip):
			m.result = confirmationSkipped
			return m, nil
		case key.Matches(msg, m.keys.Quit) || isCtrlC(msg):
			m.result = confirmationQuit
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m confirmationModel) View() tea.View {
	var total int64
	for _, candidate := range m.scan.Candidates {
		total += candidate.Size
	}

	content := []string{
		headerStyle.Render("Shed pruning"),
		"",
		fmt.Sprintf("%s will be moved to Recycle Bin. Press y/enter to confirm.", core.FormatSize(total)),
		"",
		m.list.View(),
		m.help.View(m.keys),
	}
	return tea.NewView(strings.Join(content, "\n"))
}

func (m confirmationModel) Result() confirmationResult {
	return m.result
}

type shedMonthItem string

func (item shedMonthItem) FilterValue() string {
	return string(item)
}

type shedMonthDelegate struct{}

func (shedMonthDelegate) Height() int {
	return 1
}

func (shedMonthDelegate) Spacing() int {
	return 0
}

func (shedMonthDelegate) Update(tea.Msg, *list.Model) tea.Cmd {
	return nil
}

func (shedMonthDelegate) Render(w io.Writer, _ list.Model, _ int, item list.Item) {
	_, _ = fmt.Fprint(w, listItemStyle.Render("  "+item.FilterValue()))
}

var headerStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("170")).Padding(0, 1).
	Foreground(lipgloss.Color("255"))

var listItemStyle = lipgloss.NewStyle()

func listHeight(windowHeight int) int {
	const reservedRows = 5
	height := windowHeight - reservedRows
	if height < 3 {
		return 3
	}
	return height
}

func isCtrlC(msg tea.KeyPressMsg) bool {
	pressed := msg.Key()
	return pressed.Mod&tea.ModCtrl != 0 && pressed.Code == 'c'
}
