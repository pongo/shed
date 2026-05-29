package pruning

import (
	"context"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	"shed/internal/app"
	"shed/internal/core"
)

type progressModel struct {
	ctx     context.Context
	prune   app.PruneFunc
	spinner spinner.Model
	summary core.PruneSummary
	err     error
	done    bool
}

type pruneFinishedMsg struct {
	summary core.PruneSummary
	err     error
}

func newProgressModel(ctx context.Context, prune app.PruneFunc) progressModel {
	return progressModel{
		ctx:     ctx,
		prune:   prune,
		spinner: spinner.New(spinner.WithSpinner(spinner.Line)),
	}
}

func (m progressModel) Init() tea.Cmd {
	return tea.Batch(m.runPrune, m.spinner.Tick)
}

func (m progressModel) runPrune() tea.Msg {
	summary, err := m.prune(m.ctx)
	return pruneFinishedMsg{summary: summary, err: err}
}

func (m progressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case pruneFinishedMsg:
		m.summary = msg.summary
		m.err = msg.err
		m.done = true
		return m, nil
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m progressModel) View() tea.View {
	return tea.NewView(m.spinner.View() + " Pruning Shed months")
}
