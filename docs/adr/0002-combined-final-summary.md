# Combined Final Summary

shed has separate Archive pruning and archiving phases, but their outcomes can both matter to the same run. Phase runners return outcomes and errors without printing their own final reports; the app aggregates those results and passes them to a final Bubble Tea program that renders one combined summary. This keeps phase interaction explicit in the app layer and avoids splitting final output across multiple runners, at the cost of a few shared result types for cross-phase reporting.
