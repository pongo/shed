package core

import (
	"sort"
	"strconv"
	"strings"
	"time"
)

type ArchiveMonth struct {
	Path  string
	Year  int
	Month int
}

type PruneCandidate struct {
	Month ArchiveMonth
	Size  int64
}

type PruneScanResult struct {
	Candidates []PruneCandidate
}

type PruneSummary struct {
	PrunedSize  int64
	PrunedPaths []string
	FailedPaths []string
}

func ParseArchiveMonth(path string) (ArchiveMonth, bool) {
	parts := splitPath(path)
	if len(parts) != 4 {
		return ArchiveMonth{}, false
	}
	if parts[0] != "~" || parts[1] != "Shed" {
		return ArchiveMonth{}, false
	}
	if len(parts[2]) != 4 {
		return ArchiveMonth{}, false
	}

	year, err := strconv.Atoi(parts[2])
	if err != nil || year < 0 {
		return ArchiveMonth{}, false
	}
	month, err := strconv.Atoi(parts[3])
	if err != nil || month < 1 || month > 12 {
		return ArchiveMonth{}, false
	}

	return ArchiveMonth{
		Path:  path,
		Year:  year,
		Month: month,
	}, true
}

func IsPruneEligible(month ArchiveMonth, now time.Time) bool {
	cutoff := now.AddDate(0, -6, 0)
	return monthKey(month.Year, month.Month) < monthKey(cutoff.Year(), int(cutoff.Month()))
}

func SelectPruneCandidates(candidates []PruneCandidate, now time.Time) PruneScanResult {
	result := PruneScanResult{
		Candidates: make([]PruneCandidate, 0, len(candidates)),
	}

	for _, candidate := range candidates {
		if !IsPruneEligible(candidate.Month, now) {
			continue
		}
		result.Candidates = append(result.Candidates, candidate)
	}

	sort.SliceStable(result.Candidates, func(i, j int) bool {
		left := result.Candidates[i].Month
		right := result.Candidates[j].Month
		return monthKey(left.Year, left.Month) < monthKey(right.Year, right.Month)
	})

	return result
}

func NewPruneSummary() PruneSummary {
	return PruneSummary{}
}

func (summary *PruneSummary) RecordPruned(path string, size int64) {
	summary.PrunedPaths = append(summary.PrunedPaths, path)
	summary.PrunedSize += size
}

func (summary *PruneSummary) RecordFailed(path string) {
	summary.FailedPaths = append(summary.FailedPaths, path)
}

func monthKey(year, month int) int {
	return year*100 + month
}

func splitPath(path string) []string {
	parts := strings.FieldsFunc(path, func(r rune) bool {
		return r == '\\' || r == '/'
	})
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		filtered = append(filtered, part)
	}
	return filtered
}
