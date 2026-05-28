package archiving

import (
	"context"
	"fmt"
	"io"

	tea "charm.land/bubbletea/v2"

	"shed/internal/app"
	"shed/internal/core"
)

type Runner struct {
	Input  io.Reader
	Output io.Writer
}

func (runner Runner) RunArchiving(ctx context.Context, request app.ArchivingRequest) (app.ArchivingResult, error) {
	initialModel := newModel(ctx, request)
	program := tea.NewProgram(
		initialModel,
		tea.WithContext(ctx),
		tea.WithInput(runner.Input),
		tea.WithOutput(runner.Output),
	)

	finalModel, err := program.Run()
	if err != nil {
		return app.ArchivingResult{}, err
	}
	archiving, ok := finalModel.(model)
	if !ok {
		return app.ArchivingResult{}, nil
	}
	if archiving.cancelled {
		return app.ArchivingResult{Outcome: app.ArchivingCancelled}, nil
	}
	return app.ArchivingResult{
		Outcome: app.ArchivingCompleted,
		Summary: archiving.summary,
	}, archiving.err
}

type phase int

const (
	phaseConfirming phase = iota
	phaseMoving
	phaseFinal
	phaseCancelled
)

type model struct {
	ctx          context.Context
	request      app.ArchivingRequest
	confirmation confirmationModel
	moving       movingModel
	phase        phase
	summary      core.MoveSummary
	err          error
	cancelled    bool
}

func newModel(ctx context.Context, request app.ArchivingRequest) model {
	return model{
		ctx:          ctx,
		request:      request,
		confirmation: newConfirmationModel(request.Confirmation),
		moving:       newMovingModel(ctx, request.Move, request.View),
	}
}

func (m model) Init() tea.Cmd {
	return m.confirmation.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.phase {
	case phaseConfirming:
		return m.updateConfirmation(msg)
	case phaseMoving:
		return m.updateMoving(msg)
	case phaseFinal:
		return m, nil
	case phaseCancelled:
		return m, nil
	default:
		return m, nil
	}
}

func (m model) updateConfirmation(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := m.confirmation.Update(msg)
	m.confirmation = updated.(confirmationModel)

	switch m.confirmation.Result() {
	case confirmationConfirmed:
		m.phase = phaseMoving
		return m, tea.Batch(cmd, m.moving.Init())
	case confirmationCancelled:
		m.cancelled = true
		m.phase = phaseCancelled
		return m, tea.Sequence(tea.Println("Cancelled"), tea.Quit)
	default:
		return m, cmd
	}
}

func (m model) updateMoving(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := m.moving.Update(msg)
	m.moving = updated.(movingModel)
	if m.moving.done {
		m.summary = m.moving.summary
		m.err = m.moving.err
		if m.err != nil {
			m.phase = phaseFinal
			return m, tea.Sequence(tea.Println(fmt.Sprintf("Preflight failure: %v", m.err)), tea.Quit)
		}
		m.phase = phaseFinal
		return m, tea.Sequence(tea.Println(formatFinalSummary(m.summary, m.request.View.SkippedItems)), tea.Quit)
	}
	return m, cmd
}

func (m model) View() tea.View {
	switch m.phase {
	case phaseConfirming:
		return m.confirmation.View()
	case phaseMoving:
		return m.moving.View()
	case phaseFinal:
		return tea.NewView("")
	case phaseCancelled:
		return tea.NewView("")
	default:
		return tea.NewView("Invalid archiving state")
	}
}
