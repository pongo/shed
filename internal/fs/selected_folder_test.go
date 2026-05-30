package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSelectedFolderResolverResolvesCurrentWorkingDirectory(t *testing.T) {
	cwd := t.TempDir()
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(cwd); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(original); err != nil {
			t.Fatal(err)
		}
	})

	selected, err := SelectedFolderResolver{}.Resolve(".")
	if err != nil {
		t.Fatalf("expected current working directory to resolve: %v", err)
	}
	if selected != cwd {
		t.Fatalf("expected %q, got %q", cwd, selected)
	}
}

func TestSelectedFolderResolverRejectsMissingPath(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")

	_, err := SelectedFolderResolver{}.Resolve(missing)
	if err == nil {
		t.Fatalf("expected missing path error")
	}
}

func TestSelectedFolderResolverRejectsFiles(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(file, []byte("content"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := SelectedFolderResolver{}.Resolve(file)
	if err == nil {
		t.Fatalf("expected non-folder path error")
	}
}
