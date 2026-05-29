# Combined Final Summary

shed has separate Archive pruning and archiving phases, but their outcomes can both matter to the same run. Phase runners return outcomes and errors without printing their own final reports; the app aggregates those results and passes them to a final Bubble Tea program that renders one combined summary. This keeps phase interaction explicit in the app layer and avoids splitting final output across multiple runners, at the cost of a few shared result types for cross-phase reporting.

Phase coordination remains in `internal/app`, even when the code is split across pruning, archiving, and final helper files. The implementation keeps the existing core/fs/tui layer boundaries instead of reorganizing the code into vertical feature packages, because cross-phase exit and reporting policy belongs to the application run lifecycle rather than to any single phase runner.
