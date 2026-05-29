package core

import (
	"sort"
	"time"
)

type ShedMonth struct {
	Path  string
	Year  int
	Month int
}

type PruneCandidate struct {
	Month ShedMonth
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

func IsPruneEligible(month ShedMonth, now time.Time) bool {
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
