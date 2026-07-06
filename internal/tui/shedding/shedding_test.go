package shedding

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

func TestModelDoesNotConfirmWithEmptySelection(t *testing.T) {
	model := newConfirmationModel(testRequest())

	updated, _ := model.Update(keyPress("a"))
	updated, _ = updated.(confirmationModel).Update(enterPress())
	confirmation := updated.(confirmationModel)

	if confirmation.Result() != confirmationNone {
		t.Fatalf("expected empty selection to stay unconfirmed, got %v", confirmation.Result())
	}
	if !strings.Contains(confirmation.View().Content, "0 B will be moved") {
		t.Fatalf("expected empty selection size in view, got:\n%s", confirmation.View().Content)
	}
}

func TestModelTogglesFocusedItemSelection(t *testing.T) {
	model := newConfirmationModel(testRequest())

	updated, _ := model.Update(spacePress())
	confirmation := updated.(confirmationModel)
	selected := confirmation.SelectedScanResult()

	if len(selected.StaleItems) != 1 || selected.StaleItems[0].DisplayName != "old-file.txt" {
		t.Fatalf("expected only old-file.txt selected, got %+v", selected.StaleItems)
	}
	if selected.MoveSize != 1024 {
		t.Fatalf("expected selected move size 1024, got %d", selected.MoveSize)
	}
}

func TestModelTogglesAllItemsByAggregateState(t *testing.T) {
	model := newConfirmationModel(testRequest())

	updated, _ := model.Update(keyPress("a"))
	confirmation := updated.(confirmationModel)
	if len(confirmation.SelectedScanResult().StaleItems) != 0 {
		t.Fatalf("expected all items deselected")
	}

	updated, _ = confirmation.Update(keyPress("a"))
	confirmation = updated.(confirmationModel)
	if len(confirmation.SelectedScanResult().StaleItems) != 2 {
		t.Fatalf("expected all items selected")
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

func TestSheddingModelTransitionsFromConfirmationToMoving(t *testing.T) {
	initialModel := newModel(context.Background(), testSheddingRequest())

	updated, _ := initialModel.Update(enterPress())
	shedding := updated.(model)
	view := shedding.View()

	if shedding.phase != phaseMoving {
		t.Fatalf("expected moving phase, got %v", shedding.phase)
	}
	if !strings.Contains(view.Content, "Moving items into Shed") {
		t.Fatalf("expected moving view after confirmation, got %q", view.Content)
	}
	if view.Content == "" {
		t.Fatalf("expected shedding transition not to render an empty clear view")
	}
}

func TestSheddingModelClearsManagedViewAfterCancel(t *testing.T) {
	initialModel := newModel(context.Background(), testSheddingRequest())

	updated, _ := initialModel.Update(escapePress())
	shedding := updated.(model)
	view := shedding.View()

	if shedding.phase != phaseCancelled {
		t.Fatalf("expected cancelled phase, got %v", shedding.phase)
	}
	if view.Content != "" {
		t.Fatalf("expected cancelled phase to clear managed view, got %q", view.Content)
	}
	if view.AltScreen {
		t.Fatalf("expected cancelled view to stay in the main screen buffer")
	}
}

func TestRunnerReturnsCancelledOutcomeWithoutOutputAfterEscape(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	output := new(bytes.Buffer)

	result, err := (Runner{
		Input:  strings.NewReader("\x1b"),
		Output: output,
	}).RunShedding(ctx, testSheddingRequest())

	if err != nil {
		t.Fatalf("expected runner to finish without error, got %v", err)
	}
	if result.Outcome != app.SheddingCancelled {
		t.Fatalf("expected cancelled outcome, got %v", result.Outcome)
	}
	for _, notWant := range []string{"Cancelled", "moved to", "Preflight failure:"} {
		if strings.Contains(output.String(), notWant) {
			t.Fatalf("expected no runner-owned message %q, got %q", notWant, output.String())
		}
	}
}

func TestRunnerReturnsSummaryWithoutPrintingFinalOutput(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	output := new(bytes.Buffer)
	bucket := filepath.Join("C:", "Users", "pavel", "Shed")

	result, err := (Runner{
		Input:  strings.NewReader("\r"),
		Output: output,
	}).RunShedding(ctx, app.SheddingRequest{
		Confirmation: testRequest(),
		Move: func(context.Context, core.ScanResult) (core.MoveSummary, error) {
			return core.MoveSummary{MovedSize: 10, ShedBucket: bucket}, nil
		},
		View: app.MoveViewData{},
	})

	if err != nil {
		t.Fatalf("expected runner to finish without error, got %v", err)
	}
	if result.Outcome != app.SheddingCompleted {
		t.Fatalf("expected completed outcome, got %v", result.Outcome)
	}
	if result.Summary.MovedSize != 10 || result.Summary.ShedBucket != bucket {
		t.Fatalf("unexpected returned summary: %+v", result.Summary)
	}
	for _, notWant := range []string{"Cancelled", "moved to", "Preflight failure:"} {
		if strings.Contains(output.String(), notWant) {
			t.Fatalf("expected no runner-owned message %q, got %q", notWant, output.String())
		}
	}
}

func TestRunnerReturnsMoveErrorWithoutPrinting(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	output := new(bytes.Buffer)

	result, err := (Runner{
		Input:  strings.NewReader("\r"),
		Output: output,
	}).RunShedding(ctx, app.SheddingRequest{
		Confirmation: testRequest(),
		Move: func(context.Context, core.ScanResult) (core.MoveSummary, error) {
			return core.MoveSummary{}, errors.New("selected folder unavailable")
		},
		View: app.MoveViewData{},
	})

	if err == nil {
		t.Fatalf("expected move error")
	}
	if result.Outcome != app.SheddingCompleted {
		t.Fatalf("expected completed outcome for attempted shedding, got %v", result.Outcome)
	}
	for _, notWant := range []string{"Cancelled", "moved to", "Preflight failure:"} {
		if strings.Contains(output.String(), notWant) {
			t.Fatalf("expected no runner-owned message %q, got %q", notWant, output.String())
		}
	}
}

func TestRunnerReturnsCancelledOutcomeForAllCancelKeys(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for _, input := range []string{"n", "q", "\x1b", "\x03"} {
		result, err := (Runner{
			Input:  strings.NewReader(input),
			Output: new(bytes.Buffer),
		}).RunShedding(ctx, testSheddingRequest())
		if err != nil {
			t.Fatalf("expected no error for input %q, got %v", input, err)
		}
		if result.Outcome != app.SheddingCancelled {
			t.Fatalf("expected cancelled outcome for input %q, got %v", input, result.Outcome)
		}
	}
}

func TestRunnerMovesOnlySelectedItems(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var moved core.ScanResult

	result, err := (Runner{
		Input:  strings.NewReader("a \r"),
		Output: new(bytes.Buffer),
	}).RunShedding(ctx, app.SheddingRequest{
		Confirmation: testRequest(),
		Move: func(_ context.Context, scan core.ScanResult) (core.MoveSummary, error) {
			moved = scan
			return core.MoveSummary{MovedSize: scan.MoveSize}, nil
		},
		View: app.MoveViewData{},
	})

	if err != nil {
		t.Fatalf("expected runner to finish without error, got %v", err)
	}
	if result.Outcome != app.SheddingCompleted {
		t.Fatalf("expected completed outcome, got %v", result.Outcome)
	}
	if len(moved.StaleItems) != 1 || moved.StaleItems[0].DisplayName != "old-folder" {
		t.Fatalf("expected only old-folder to move, got %+v", moved.StaleItems)
	}
	if moved.MoveSize != 2048 {
		t.Fatalf("expected selected move size 2048, got %d", moved.MoveSize)
	}
}

func TestSheddingModelClearsManagedViewAfterMove(t *testing.T) {
	bucket := filepath.Join("C:", "Users", "pavel", "Shed")
	initialModel := newModel(context.Background(), testSheddingRequest())
	initialModel.phase = phaseMoving

	updated, _ := initialModel.Update(moveFinishedMsg{summary: core.MoveSummary{MovedSize: 10, ShedBucket: bucket}})
	shedding := updated.(model)
	view := shedding.View().Content

	if shedding.phase != phaseFinal {
		t.Fatalf("expected final phase, got %v", shedding.phase)
	}
	if view != "" {
		t.Fatalf("expected final phase to clear managed view, got %q", view)
	}
}

func TestSheddingModelRendersPreflightFailureAfterMoveError(t *testing.T) {
	initialModel := newModel(context.Background(), testSheddingRequest())
	initialModel.phase = phaseMoving

	updated, _ := initialModel.Update(moveFinishedMsg{err: errors.New("selected folder unavailable")})
	shedding := updated.(model)
	view := shedding.View().Content

	if shedding.phase != phasePreflightFailure {
		t.Fatalf("expected preflight failure phase, got %v", shedding.phase)
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
	if !strings.Contains(view, "Skipped items: 2.\n\n"+currentListItemStyle.Render("  [x] old-folder")) {
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

	for _, want := range []string{"  [x] old-folder", "  [x] old-file.txt"} {
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

func TestListItemsRenderWithKindStyles(t *testing.T) {
	model := newConfirmationModel(testRequest())
	model.list.Select(1)
	view := model.View().Content

	for _, want := range []string{
		folderListItemStyle.Render("  [x] old-folder"),
		currentListItemStyle.Render("  [x] old-file.txt"),
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected styled list item %q, got:\n%s", want, view)
		}
	}
}

func TestListItemsRenderUnselectedItemsGray(t *testing.T) {
	model := newConfirmationModel(testRequest())
	model.list.Select(1)

	updated, _ := model.Update(spacePress())
	view := updated.(confirmationModel).View().Content

	if !strings.Contains(view, currentListItemStyle.Render("  [ ] old-file.txt")) {
		t.Fatalf("expected current unselected item to render with cursor color, got:\n%s", view)
	}
	if !strings.Contains(view, "2 KB will be moved") {
		t.Fatalf("expected selected size in summary, got:\n%s", view)
	}
}

func TestListItemsRenderUnselectedInactiveItemsGray(t *testing.T) {
	model := newConfirmationModel(testRequest())

	updated, _ := model.Update(spacePress())
	confirmation := updated.(confirmationModel)
	confirmation.list.Select(1)
	view := confirmation.View().Content

	if !strings.Contains(view, unselectedListItemStyle.Render("  [ ] old-folder")) {
		t.Fatalf("expected inactive unselected item to render gray, got:\n%s", view)
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
	model := newMovingModel(context.Background(), func(context.Context, core.ScanResult) (core.MoveSummary, error) {
		return core.MoveSummary{}, nil
	}, app.MoveViewData{})

	view := model.View().Content
	if !strings.Contains(view, "Moving items into Shed") {
		t.Fatalf("expected moving view, got %q", view)
	}
}

func testSheddingRequest() app.SheddingRequest {
	return app.SheddingRequest{
		Confirmation: testRequest(),
		Move: func(context.Context, core.ScanResult) (core.MoveSummary, error) {
			return core.MoveSummary{}, nil
		},
		View: app.MoveViewData{},
	}
}

func testRequest() app.ConfirmationRequest {
	return app.ConfirmationRequest{
		SelectedFolder:    `C:\Users\pavel\Downloads`,
		HeaderTitle:       "Downloads",
		CompactShedBucket: filepath.Join("~", "Shed", "2026", "05", "Downloads"),
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

func spacePress() tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: tea.KeySpace})
}

func escapePress() tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape})
}

func ctrlCPress() tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Text: "c", Code: 'c', Mod: tea.ModCtrl})
}
