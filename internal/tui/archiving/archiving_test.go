package archiving

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"shed/internal/app"
	"shed/internal/core"
)

func TestModelConfirmsWithYAndEnter(t *testing.T) {
	for _, msg := range []tea.KeyPressMsg{keyPress("y"), enterPress()} {
		model := newConfirmationModel(testRequest())

		updated, _ := model.Update(msg)
		result := updated.(confirmationModel).Result()

		if result != confirmationConfirmed {
			t.Fatalf("expected confirm result, got %v", result)
		}
	}
}

func TestModelCancelsWithConfiguredKeys(t *testing.T) {
	for _, msg := range []tea.KeyPressMsg{keyPress("n"), keyPress("q"), escapePress(), ctrlCPress()} {
		model := newConfirmationModel(testRequest())

		updated, _ := model.Update(msg)
		result := updated.(confirmationModel).Result()

		if result != confirmationCancelled {
			t.Fatalf("expected cancel result for %q, got %v", msg.String(), result)
		}
	}
}

func TestArchivingModelTransitionsFromConfirmationToMoving(t *testing.T) {
	initialModel := newModel(context.Background(), testArchivingRequest())

	updated, _ := initialModel.Update(enterPress())
	archiving := updated.(model)
	view := archiving.View()

	if archiving.phase != phaseMoving {
		t.Fatalf("expected moving phase, got %v", archiving.phase)
	}
	if !strings.Contains(view.Content, "Moving items into Archive") {
		t.Fatalf("expected moving view after confirmation, got %q", view.Content)
	}
	if view.Content == "" {
		t.Fatalf("expected archiving transition not to render an empty clear view")
	}
}

func TestArchivingModelClearsManagedViewAfterCancel(t *testing.T) {
	initialModel := newModel(context.Background(), testArchivingRequest())

	updated, _ := initialModel.Update(escapePress())
	archiving := updated.(model)
	view := archiving.View()

	if archiving.phase != phaseCancelled {
		t.Fatalf("expected cancelled phase, got %v", archiving.phase)
	}
	if view.Content != "" {
		t.Fatalf("expected cancelled phase to clear managed view, got %q", view.Content)
	}
	if view.AltScreen {
		t.Fatalf("expected cancelled view to stay in the main screen buffer")
	}
}

func TestRunnerRendersCancelledOutputAfterEscape(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	output := new(bytes.Buffer)

	result, err := (Runner{
		Input:  strings.NewReader("\x1b"),
		Output: output,
	}).RunArchiving(ctx, testArchivingRequest())

	if err != nil {
		t.Fatalf("expected runner to finish without error, got %v", err)
	}
	if result.Outcome != app.ArchivingCancelled {
		t.Fatalf("expected cancelled outcome, got %v", result.Outcome)
	}
	if !strings.Contains(output.String(), "Cancelled") {
		t.Fatalf("expected output to contain Cancelled, got %q", output.String())
	}
}

func TestRunnerRendersFinalSummaryOutputAfterMove(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	output := new(bytes.Buffer)
	bucket := filepath.Join("C:", "Users", "pavel", "Shed")

	result, err := (Runner{
		Input:  strings.NewReader("\r"),
		Output: output,
	}).RunArchiving(ctx, app.ArchivingRequest{
		Confirmation: testRequest(),
		Move: func(context.Context) (core.MoveSummary, error) {
			return core.MoveSummary{MovedSize: 10, ArchiveBucket: bucket}, nil
		},
		View: app.MoveViewData{},
	})

	if err != nil {
		t.Fatalf("expected runner to finish without error, got %v", err)
	}
	if result.Outcome != app.ArchivingCompleted {
		t.Fatalf("expected completed outcome, got %v", result.Outcome)
	}
	if !strings.Contains(output.String(), "10 B moved to "+bucket) {
		t.Fatalf("expected output to contain final summary, got %q", output.String())
	}
}

func TestRunnerRendersPreflightFailureOutputAfterMoveError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	output := new(bytes.Buffer)

	result, err := (Runner{
		Input:  strings.NewReader("\r"),
		Output: output,
	}).RunArchiving(ctx, app.ArchivingRequest{
		Confirmation: testRequest(),
		Move: func(context.Context) (core.MoveSummary, error) {
			return core.MoveSummary{}, errors.New("selected folder unavailable")
		},
		View: app.MoveViewData{},
	})

	if err == nil {
		t.Fatalf("expected move error")
	}
	if result.Outcome != app.ArchivingCompleted {
		t.Fatalf("expected completed outcome for attempted archiving, got %v", result.Outcome)
	}
	if !strings.Contains(output.String(), "Preflight failure: selected folder unavailable") {
		t.Fatalf("expected output to contain preflight failure, got %q", output.String())
	}
}

func TestArchivingModelClearsManagedViewAfterMove(t *testing.T) {
	bucket := filepath.Join("C:", "Users", "pavel", "Shed")
	initialModel := newModel(context.Background(), testArchivingRequest())
	initialModel.phase = phaseMoving

	updated, _ := initialModel.Update(moveFinishedMsg{summary: core.MoveSummary{MovedSize: 10, ArchiveBucket: bucket}})
	archiving := updated.(model)
	view := archiving.View().Content

	if archiving.phase != phaseFinal {
		t.Fatalf("expected final phase, got %v", archiving.phase)
	}
	if view != "" {
		t.Fatalf("expected final phase to clear managed view, got %q", view)
	}
}

func TestArchivingModelRendersPreflightFailureAfterMoveError(t *testing.T) {
	initialModel := newModel(context.Background(), testArchivingRequest())
	initialModel.phase = phaseMoving

	updated, _ := initialModel.Update(moveFinishedMsg{err: errors.New("selected folder unavailable")})
	archiving := updated.(model)
	view := archiving.View().Content

	if archiving.phase != phasePreflightFailure {
		t.Fatalf("expected preflight failure phase, got %v", archiving.phase)
	}
	if view != "" {
		t.Fatalf("expected preflight failure phase to clear managed view, got %q", view)
	}
}

func TestViewRendersConfirmationText(t *testing.T) {
	view := newConfirmationModel(testRequest()).View().Content

	for _, want := range []string{"Downloads", "3 KB will be moved to " + filepath.Join("~", "Shed", "2026", "05", "Downloads"), "Press y/enter to confirm.", "Skipped items: 2.", "old-folder", "old-file.txt", "y/enter", "n/q/esc"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected view to contain %q, got:\n%s", want, view)
		}
	}

	if !strings.Contains(view, "\n\n3 KB will be moved") {
		t.Fatalf("expected blank space after header, got:\n%s", view)
	}
	if !strings.Contains(view, "Skipped items: 2.\n\n"+listItemStyle.Render("  old-folder")) {
		t.Fatalf("expected blank space before list, got:\n%s", view)
	}
}

func TestViewDoesNotUseAltScreen(t *testing.T) {
	view := newConfirmationModel(testRequest()).View()

	if view.AltScreen {
		t.Fatalf("expected confirmation view to stay in the main screen buffer")
	}
}

func TestViewUsesDisplayNamesOnly(t *testing.T) {
	view := newConfirmationModel(testRequest()).View().Content

	for _, want := range []string{"  old-folder", "  old-file.txt"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected indented list item %q, got:\n%s", want, view)
		}
	}

	for _, notWant := range []string{`C:\Users\pavel\Downloads\old-folder`, "2048", "old folder description"} {
		if strings.Contains(view, notWant) {
			t.Fatalf("expected view not to contain %q, got:\n%s", notWant, view)
		}
	}
}

func TestListItemsRenderMuted(t *testing.T) {
	view := newConfirmationModel(testRequest()).View().Content
	want := listItemStyle.Render("  old-folder")

	if !strings.Contains(view, want) {
		t.Fatalf("expected muted list item %q, got:\n%s", want, view)
	}
}

func TestWindowResizeUpdatesListHeight(t *testing.T) {
	model := newConfirmationModel(testRequest())

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	resized := updated.(confirmationModel)

	if resized.ListHeight() != 19 {
		t.Fatalf("expected list height 19, got %d", resized.ListHeight())
	}
}

func TestMovingModelRendersSpinnerState(t *testing.T) {
	model := newMovingModel(context.Background(), func(context.Context) (core.MoveSummary, error) {
		return core.MoveSummary{}, nil
	}, app.MoveViewData{})

	view := model.View().Content
	if !strings.Contains(view, "Moving items into Archive") {
		t.Fatalf("expected moving view, got %q", view)
	}
}

func testArchivingRequest() app.ArchivingRequest {
	return app.ArchivingRequest{
		Confirmation: testRequest(),
		Move: func(context.Context) (core.MoveSummary, error) {
			return core.MoveSummary{}, nil
		},
		View: app.MoveViewData{},
	}
}

func testRequest() app.ConfirmationRequest {
	return app.ConfirmationRequest{
		SelectedFolder:       `C:\Users\pavel\Downloads`,
		HeaderTitle:          "Downloads",
		CompactArchiveBucket: filepath.Join("~", "Shed", "2026", "05", "Downloads"),
		ScanResult: core.ScanResult{
			StaleItems: []core.StaleItem{
				{DisplayName: "old-folder", Path: `C:\Users\pavel\Downloads\old-folder`, Kind: core.FolderItem, MoveSize: 2048},
				{DisplayName: "old-file.txt", Path: `C:\Users\pavel\Downloads\old-file.txt`, Kind: core.FileItem, MoveSize: 1024},
			},
			SkippedItems: []core.SkippedItem{
				{Path: `C:\Users\pavel\Downloads\bad-one`},
				{Path: `C:\Users\pavel\Downloads\bad-two`},
			},
			MoveSize: 3072,
		},
	}
}

func keyPress(text string) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Text: text, Code: []rune(text)[0]})
}

func enterPress() tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
}

func escapePress() tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape})
}

func ctrlCPress() tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Text: "c", Code: 'c', Mod: tea.ModCtrl})
}
