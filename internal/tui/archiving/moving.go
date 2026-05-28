package archiving

import (
	"context"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	"shed/internal/app"
	"shed/internal/core"
)

type movingModel struct {
	ctx     context.Context
	move    app.MoveFunc
	view    app.MoveViewData
	spinner spinner.Model
	summary core.MoveSummary
	err     error
	done    bool
}

type moveFinishedMsg struct {
	summary core.MoveSummary
	err     error
}

func newMovingModel(ctx context.Context, move app.MoveFunc, view app.MoveViewData) movingModel {
	return movingModel{
		ctx:     ctx,
		move:    move,
		view:    view,
		spinner: spinner.New(spinner.WithSpinner(spinner.Line)),
	}
}

func (m movingModel) Init() tea.Cmd {
	return tea.Batch(m.runMove, m.spinner.Tick)
}

func (m movingModel) runMove() tea.Msg {
	summary, err := m.move(m.ctx)
	return moveFinishedMsg{summary: summary, err: err}
}

func (m movingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case moveFinishedMsg:
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

func (m movingModel) View() tea.View {
	return tea.NewView(m.spinner.View() + " Moving items into Archive")
}
