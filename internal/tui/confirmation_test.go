package tui

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

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

func TestViewClearsAfterConfirmationResult(t *testing.T) {
	for _, msg := range []tea.KeyPressMsg{enterPress(), keyPress("n")} {
		model := newConfirmationModel(testRequest())

		updated, _ := model.Update(msg)
		view := updated.(confirmationModel).View()

		if view.Content != "" {
			t.Fatalf("expected empty view after result, got:\n%s", view.Content)
		}
		if view.AltScreen {
			t.Fatalf("expected cleared view to stay in the main screen buffer")
		}
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

func TestMovingModelRendersFinalSummaryAfterMove(t *testing.T) {
	bucket := filepath.Join("C:", "Users", "pavel", "Shed")
	model := newMovingModel(context.Background(), func(context.Context) (core.MoveSummary, error) {
		return core.MoveSummary{MovedSize: 10, ArchiveBucket: bucket}, nil
	}, app.MoveViewData{})

	updated, _ := model.Update(moveFinishedMsg{summary: core.MoveSummary{MovedSize: 10, ArchiveBucket: bucket}})
	view := updated.(movingModel).View().Content

	if !strings.Contains(view, "10 B moved to "+bucket) {
		t.Fatalf("expected final move summary, got %q", view)
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
