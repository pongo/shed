package fs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/hymkor/trash-go"

	"shed/internal/core"
)

type Pruner struct {
	ArchiveRoot string
	Now         func() time.Time
	ReadDir     func(path string) ([]os.DirEntry, error)
	Size        func(path string) (int64, error)
	Throw       func(path string) error
}

func NewPruner(archiveRoot string) Pruner {
	return Pruner{
		ArchiveRoot: archiveRoot,
		Now:         time.Now,
		ReadDir:     os.ReadDir,
		Size:        RecursiveSize,
		Throw: func(path string) error {
			return trash.Throw(path)
		},
	}
}

func (pruner Pruner) Scan(ctx context.Context) (core.PruneScanResult, error) {
	pruner = pruner.withDefaults()

	if _, err := os.Stat(pruner.ArchiveRoot); err != nil {
		if os.IsNotExist(err) {
			return core.PruneScanResult{}, nil
		}
		return core.PruneScanResult{}, err
	}

	yearEntries, err := pruner.ReadDir(pruner.ArchiveRoot)
	if err != nil {
		return core.PruneScanResult{}, err
	}

	candidates := make([]core.PruneCandidate, 0)
	scanErrors := make([]error, 0)

	for _, yearEntry := range yearEntries {
		if err := ctx.Err(); err != nil {
			return core.PruneScanResult{}, err
		}
		if !yearEntry.IsDir() {
			continue
		}

		year, ok := parseYear(yearEntry.Name())
		if !ok {
			continue
		}

		yearPath := filepath.Join(pruner.ArchiveRoot, yearEntry.Name())
		monthEntries, err := pruner.ReadDir(yearPath)
		if err != nil {
			scanErrors = append(scanErrors, fmt.Errorf("scan year %s: %w", yearPath, err))
			continue
		}

		for _, monthEntry := range monthEntries {
			if err := ctx.Err(); err != nil {
				return core.PruneScanResult{}, err
			}
			if !monthEntry.IsDir() {
				continue
			}

			month, ok := parseMonth(monthEntry.Name())
			if !ok {
				continue
			}

			monthPath := filepath.Join(yearPath, monthEntry.Name())
			size, err := pruner.Size(monthPath)
			if err != nil {
				scanErrors = append(scanErrors, fmt.Errorf("calculate size for %s: %w", monthPath, err))
				continue
			}

			candidates = append(candidates, core.PruneCandidate{
				Month: core.ArchiveMonth{
					Path:  filepath.Join("~", "Shed", yearEntry.Name(), monthEntry.Name()),
					Year:  year,
					Month: month,
				},
				Size: size,
			})
		}
	}

	result := core.SelectPruneCandidates(candidates, pruner.Now())
	if len(scanErrors) > 0 {
		return result, errors.Join(scanErrors...)
	}
	return result, nil
}

func (pruner Pruner) Prune(ctx context.Context, scan core.PruneScanResult) (core.PruneSummary, error) {
	pruner = pruner.withDefaults()

	summary := core.NewPruneSummary()
	for _, candidate := range scan.Candidates {
		if err := ctx.Err(); err != nil {
			return summary, err
		}

		monthPath := filepath.Join(pruner.ArchiveRoot, fmt.Sprintf("%04d", candidate.Month.Year), fmt.Sprintf("%02d", candidate.Month.Month))
		if err := pruner.Throw(monthPath); err != nil {
			summary.RecordFailed(candidate.Month.Path)
			continue
		}

		summary.RecordPruned(candidate.Month.Path, candidate.Size)
		pruner.cleanupEmptyYear(filepath.Dir(monthPath))
	}

	return summary, nil
}

func (pruner Pruner) cleanupEmptyYear(yearPath string) {
	entries, err := pruner.ReadDir(yearPath)
	if err != nil || len(entries) > 0 {
		return
	}
	_ = pruner.Throw(yearPath)
}

func (pruner Pruner) withDefaults() Pruner {
	if pruner.Now == nil {
		pruner.Now = time.Now
	}
	if pruner.ReadDir == nil {
		pruner.ReadDir = os.ReadDir
	}
	if pruner.Size == nil {
		pruner.Size = RecursiveSize
	}
	if pruner.Throw == nil {
		pruner.Throw = func(path string) error {
			return trash.Throw(path)
		}
	}
	return pruner
}

func parseYear(name string) (int, bool) {
	if len(name) != 4 {
		return 0, false
	}
	year, err := strconv.Atoi(name)
	if err != nil || year < 0 {
		return 0, false
	}
	return year, true
}

func parseMonth(name string) (int, bool) {
	if len(name) != 2 {
		return 0, false
	}
	month, err := strconv.Atoi(name)
	if err != nil || month < 1 || month > 12 {
		return 0, false
	}
	return month, true
}
