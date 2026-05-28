package tui

import (
	"context"
	"io"

	tea "charm.land/bubbletea/v2"

	"shed/internal/app"
)

type Confirmer struct {
	Input  io.Reader
	Output io.Writer
}

func (confirmer Confirmer) Confirm(ctx context.Context, request app.ConfirmationRequest) (app.ConfirmationOutcome, error) {
	model := NewModel(request)
	program := tea.NewProgram(
		model,
		tea.WithContext(ctx),
		tea.WithInput(confirmer.Input),
		tea.WithOutput(confirmer.Output),
	)

	finalModel, err := program.Run()
	if err != nil {
		return app.ConfirmationCancelled, err
	}

	if resultModel, ok := finalModel.(Model); ok && resultModel.Result() == ResultConfirm {
		return app.ConfirmationConfirmed, nil
	}
	return app.ConfirmationCancelled, nil
}
