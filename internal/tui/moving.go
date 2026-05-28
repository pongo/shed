package tui

import (
	"context"
	"io"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	"shed/internal/app"
	"shed/internal/core"
)

type MovingRunner struct {
	Input  io.Reader
	Output io.Writer
}

func (runner MovingRunner) RunMoving(ctx context.Context, move app.MoveFunc, view app.MoveViewData) (core.MoveSummary, error) {
	model := newMovingModel(ctx, move, view)
	program := tea.NewProgram(
		model,
		tea.WithContext(ctx),
		tea.WithInput(runner.Input),
		tea.WithOutput(runner.Output),
	)

	finalModel, err := program.Run()
	if err != nil {
		return core.MoveSummary{}, err
	}
	moving, ok := finalModel.(movingModel)
	if !ok {
		return core.MoveSummary{}, nil
	}
	return moving.summary, moving.err
}

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
		return m, tea.Quit
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m movingModel) View() tea.View {
	if m.done {
		if m.err != nil {
			return tea.NewView("")
		}
		return finalSummaryView(m.summary, m.view.SkippedItems)
	}
	return tea.NewView(m.spinner.View() + " Moving items into Archive")
}
