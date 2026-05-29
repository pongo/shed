package core

import "testing"

func TestDecideShedMoveUsesNumberedSuffixForFileConflict(t *testing.T) {
	decision := DecideShedMove(
		ShedMoveCandidate{Name: "report.pdf", Kind: FileItem},
		[]ShedMoveEntry{{Name: "report.pdf", Kind: FileItem}},
	)

	if decision.Action != MoveIntoShed {
		t.Fatalf("expected move action, got %v", decision.Action)
	}
	if decision.TargetName != "report (1).pdf" {
		t.Fatalf("expected numbered target, got %q", decision.TargetName)
	}
}

func TestDecideShedMoveMergesFolderConflict(t *testing.T) {
	decision := DecideShedMove(
		ShedMoveCandidate{Name: "project", Kind: FolderItem},
		[]ShedMoveEntry{{Name: "project", Kind: FolderItem}},
	)

	if decision.Action != MergeIntoExistingFolder {
		t.Fatalf("expected merge action, got %v", decision.Action)
	}
	if decision.TargetName != "project" {
		t.Fatalf("expected existing folder target, got %q", decision.TargetName)
	}
}

func TestDecideShedMoveRenamesFolderWhenExistingNameIsFile(t *testing.T) {
	decision := DecideShedMove(
		ShedMoveCandidate{Name: "project", Kind: FolderItem},
		[]ShedMoveEntry{{Name: "project", Kind: FileItem}},
	)

	if decision.Action != MoveIntoShed {
		t.Fatalf("expected move action, got %v", decision.Action)
	}
	if decision.TargetName != "project (1)" {
		t.Fatalf("expected numbered folder target, got %q", decision.TargetName)
	}
}

func TestMoveSummaryRecordsMovedSizeAndSpecificFailedPaths(t *testing.T) {
	summary := NewMoveSummary(`C:\Users\pavel\Shed\2026\05\Downloads`)

	summary.RecordMoved(12)
	summary.RecordFailed(`C:\Users\pavel\Downloads\project\nested\file.txt`)

	if summary.MovedSize != 12 {
		t.Fatalf("expected moved size 12, got %d", summary.MovedSize)
	}
	if len(summary.FailedPaths) != 1 {
		t.Fatalf("expected one failed path, got %d", len(summary.FailedPaths))
	}
	if summary.FailedPaths[0] != `C:\Users\pavel\Downloads\project\nested\file.txt` {
		t.Fatalf("expected nested failed path, got %q", summary.FailedPaths[0])
	}
}
