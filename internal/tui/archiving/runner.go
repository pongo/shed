package archiving

import (
	"context"
	"fmt"
	"io"
	"time"

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
	output := runner.Output
	if output == nil {
		output = io.Discard
	}
	program := tea.NewProgram(
		initialModel,
		tea.WithContext(ctx),
		tea.WithInput(runner.Input),
		tea.WithOutput(output),
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
		fmt.Fprintln(output, "Cancelled")
		return app.ArchivingResult{Outcome: app.ArchivingCancelled}, nil
	}
	if archiving.err != nil {
		fmt.Fprintf(output, "Preflight failure: %v\n", archiving.err)
	}
	if archiving.err == nil {
		fmt.Fprint(output, formatFinalSummary(archiving.summary, request.View.SkippedItems))
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
	phasePreflightFailure
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

type quitMsg struct{}

const finalRenderDelay = 50 * time.Millisecond

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
	if _, ok := msg.(quitMsg); ok {
		return m, tea.Quit
	}

	switch m.phase {
	case phaseConfirming:
		return m.updateConfirmation(msg)
	case phaseMoving:
		return m.updateMoving(msg)
	case phaseFinal:
		return m, nil
	case phaseCancelled:
		return m, nil
	case phasePreflightFailure:
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
		return m, quitAfterClearRender
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
			m.phase = phasePreflightFailure
			return m, quitAfterClearRender
		}
		m.phase = phaseFinal
		return m, quitAfterClearRender
	}
	return m, cmd
}

func quitAfterClearRender() tea.Msg {
	time.Sleep(finalRenderDelay)
	return quitMsg{}
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
	case phasePreflightFailure:
		return tea.NewView("")
	default:
		return tea.NewView("Invalid archiving state")
	}
}
