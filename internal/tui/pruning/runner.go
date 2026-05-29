package pruning

import (
	"context"
	"io"

	tea "charm.land/bubbletea/v2"

	"shed/internal/app"
	"shed/internal/core"
)

type Runner struct {
	Input  io.Reader
	Output io.Writer
}

type phase int

const (
	phaseConfirming phase = iota
	phasePruning
	phaseDone
	phaseSkipped
	phaseQuit
)

type model struct {
	ctx          context.Context
	request      app.PruningRequest
	confirmation confirmationModel
	pruning      progressModel
	phase        phase
	outcome      app.PruningOutcome
	summary      core.PruneSummary
	err          error
}

func (runner Runner) RunPruning(ctx context.Context, request app.PruningRequest) (app.PruningResult, error) {
	output := runner.Output
	if output == nil {
		output = io.Discard
	}

	initialModel := newModel(ctx, request)
	program := tea.NewProgram(
		initialModel,
		tea.WithContext(ctx),
		tea.WithInput(runner.Input),
		tea.WithOutput(output),
	)

	finalModel, err := program.Run()
	if err != nil {
		return app.PruningResult{}, err
	}

	m, ok := finalModel.(model)
	if !ok {
		return app.PruningResult{}, nil
	}

	return app.PruningResult{
		Outcome: m.outcome,
		Summary: m.summary,
	}, m.err
}

func newModel(ctx context.Context, request app.PruningRequest) model {
	return model{
		ctx:          ctx,
		request:      request,
		confirmation: newConfirmationModel(request.Scan),
		pruning:      newProgressModel(ctx, request.Prune),
		outcome:      app.PruningSkipped,
	}
}

func (m model) Init() tea.Cmd {
	return m.confirmation.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.phase {
	case phaseConfirming:
		return m.updateConfirmation(msg)
	case phasePruning:
		return m.updatePruning(msg)
	default:
		return m, nil
	}
}

func (m model) updateConfirmation(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := m.confirmation.Update(msg)
	m.confirmation = updated.(confirmationModel)

	switch m.confirmation.Result() {
	case confirmationConfirmed:
		m.phase = phasePruning
		m.outcome = app.PruningConfirmed
		return m, tea.Batch(cmd, m.pruning.Init())
	case confirmationSkipped:
		m.phase = phaseSkipped
		m.outcome = app.PruningSkipped
		return m, tea.Quit
	case confirmationQuit:
		m.phase = phaseQuit
		m.outcome = app.PruningQuit
		return m, tea.Quit
	default:
		return m, cmd
	}
}

func (m model) updatePruning(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := m.pruning.Update(msg)
	m.pruning = updated.(progressModel)
	if m.pruning.done {
		m.summary = m.pruning.summary
		m.err = m.pruning.err
		m.phase = phaseDone
		return m, tea.Quit
	}
	return m, cmd
}

func (m model) View() tea.View {
	switch m.phase {
	case phaseConfirming:
		return m.confirmation.View()
	case phasePruning:
		return m.pruning.View()
	default:
		return tea.NewView("")
	}
}
