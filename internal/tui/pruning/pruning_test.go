package pruning

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

func TestConfirmationKeys(t *testing.T) {
	for _, msg := range []tea.KeyPressMsg{keyPress("y"), enterPress()} {
		model := newConfirmationModel(testScan())
		updated, _ := model.Update(msg)
		if updated.(confirmationModel).Result() != confirmationConfirmed {
			t.Fatalf("expected confirmed for key %q", msg.String())
		}
	}
}

func TestSkipKeys(t *testing.T) {
	for _, msg := range []tea.KeyPressMsg{keyPress("n"), escapePress()} {
		model := newConfirmationModel(testScan())
		updated, _ := model.Update(msg)
		if updated.(confirmationModel).Result() != confirmationSkipped {
			t.Fatalf("expected skipped for key %q", msg.String())
		}
	}
}

func TestQuitKeys(t *testing.T) {
	for _, msg := range []tea.KeyPressMsg{keyPress("q"), ctrlCPress()} {
		model := newConfirmationModel(testScan())
		updated, _ := model.Update(msg)
		if updated.(confirmationModel).Result() != confirmationQuit {
			t.Fatalf("expected quit for key %q", msg.String())
		}
	}
}

func TestConfirmationViewShowsTotalAndMonths(t *testing.T) {
	view := newConfirmationModel(testScan()).View().Content

	for _, want := range []string{
		"Archive pruning",
		"3 KB will be moved to Recycle Bin. Press y/enter to confirm.",
		filepath.Join("~", "Shed", "2024", "01"),
		filepath.Join("~", "Shed", "2024", "03"),
		"y/enter",
		"n/esc",
		"q/ctrl+c",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected view to contain %q, got:\n%s", want, view)
		}
	}
}

func TestProgressViewShownAfterConfirm(t *testing.T) {
	m := newModel(context.Background(), testRequest(func(context.Context) (core.PruneSummary, error) {
		return core.PruneSummary{}, nil
	}))

	updated, _ := m.Update(enterPress())
	pruning := updated.(model)
	view := pruning.View().Content

	if pruning.phase != phasePruning {
		t.Fatalf("expected pruning phase, got %v", pruning.phase)
	}
	if !strings.Contains(view, "Pruning Archive months") {
		t.Fatalf("expected progress content, got %q", view)
	}
}

func TestRunnerReturnsSkippedOutcome(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result, err := (Runner{
		Input:  strings.NewReader("\x1b"),
		Output: new(bytes.Buffer),
	}).RunPruning(ctx, testRequest(func(context.Context) (core.PruneSummary, error) {
		t.Fatalf("prune should not be called on skip")
		return core.PruneSummary{}, nil
	}))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Outcome != app.PruningSkipped {
		t.Fatalf("expected skipped outcome, got %v", result.Outcome)
	}
}

func TestRunnerReturnsQuitOutcome(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result, err := (Runner{
		Input:  strings.NewReader("q"),
		Output: new(bytes.Buffer),
	}).RunPruning(ctx, testRequest(func(context.Context) (core.PruneSummary, error) {
		t.Fatalf("prune should not be called on quit")
		return core.PruneSummary{}, nil
	}))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Outcome != app.PruningQuit {
		t.Fatalf("expected quit outcome, got %v", result.Outcome)
	}
}

func TestRunnerReturnsSummaryAndError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	wantErr := errors.New("prune failed")
	wantSummary := core.PruneSummary{
		PrunedSize:  10,
		PrunedPaths: []string{filepath.Join("~", "Shed", "2024", "01")},
	}

	result, err := (Runner{
		Input:  strings.NewReader("\r"),
		Output: new(bytes.Buffer),
	}).RunPruning(ctx, testRequest(func(context.Context) (core.PruneSummary, error) {
		return wantSummary, wantErr
	}))
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected prune error %v, got %v", wantErr, err)
	}
	if result.Outcome != app.PruningConfirmed {
		t.Fatalf("expected confirmed outcome, got %v", result.Outcome)
	}
	if result.Summary.PrunedSize != 10 || len(result.Summary.PrunedPaths) != 1 {
		t.Fatalf("unexpected summary: %+v", result.Summary)
	}
}

func testRequest(prune app.PruneFunc) app.PruningRequest {
	return app.PruningRequest{
		Scan:  testScan(),
		Prune: prune,
	}
}

func testScan() core.PruneScanResult {
	return core.PruneScanResult{
		Candidates: []core.PruneCandidate{
			{Month: core.ArchiveMonth{Path: filepath.Join("~", "Shed", "2024", "01"), Year: 2024, Month: 1}, Size: 2048},
			{Month: core.ArchiveMonth{Path: filepath.Join("~", "Shed", "2024", "03"), Year: 2024, Month: 3}, Size: 1024},
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
