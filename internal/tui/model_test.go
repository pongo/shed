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
		model := NewModel(testRequest())

		updated, _ := model.Update(msg)
		result := updated.(Model).Result()

		if result != ResultConfirm {
			t.Fatalf("expected confirm result, got %v", result)
		}
	}
}

func TestModelCancelsWithConfiguredKeys(t *testing.T) {
	for _, msg := range []tea.KeyPressMsg{keyPress("n"), keyPress("q"), escapePress(), ctrlCPress()} {
		model := NewModel(testRequest())

		updated, _ := model.Update(msg)
		result := updated.(Model).Result()

		if result != ResultCancel {
			t.Fatalf("expected cancel result for %q, got %v", msg.String(), result)
		}
	}
}

func TestViewRendersConfirmationText(t *testing.T) {
	view := NewModel(testRequest()).View().Content

	for _, want := range []string{"Downloads", "Move size: 3 KB", filepath.Join("~", "Shed", "2026", "05", "Downloads"), "Skipped: 2", "old-folder", "old-file.txt", "y/enter", "n/q/esc"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected view to contain %q, got:\n%s", want, view)
		}
	}
}

func TestViewUsesDisplayNamesOnly(t *testing.T) {
	view := NewModel(testRequest()).View().Content

	for _, notWant := range []string{`C:\Users\pavel\Downloads\old-folder`, "2048", "old folder description"} {
		if strings.Contains(view, notWant) {
			t.Fatalf("expected view not to contain %q, got:\n%s", notWant, view)
		}
	}
}

func TestWindowResizeUpdatesListHeight(t *testing.T) {
	model := NewModel(testRequest())

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	resized := updated.(Model)

	if resized.ListHeight() != 17 {
		t.Fatalf("expected list height 17, got %d", resized.ListHeight())
	}
}

func TestMovingModelRendersSpinnerState(t *testing.T) {
	model := newMovingModel(context.Background(), func(context.Context) (core.MoveSummary, error) {
		return core.MoveSummary{}, nil
	})

	view := model.View().Content
	if !strings.Contains(view, "Moving items into Archive") {
		t.Fatalf("expected moving view, got %q", view)
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
