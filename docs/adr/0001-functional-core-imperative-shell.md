# Functional Core, Imperative Shell

shed keeps filesystem decisions in a functional core and leaves real Windows filesystem access and Bubble Tea interaction in an imperative shell. This is a deliberate trade-off: the program has more explicit models and adapters than a small CLI strictly needs, but most behavior can be tested without touching real files while the thin adapters get focused integration coverage.

Core tests use in-memory models for stale rules, hidden rules, sorting, Shed bucket paths, conflict resolution, merge planning, and summaries. Adapter tests may use temporary directories for Windows-specific filesystem behavior such as hidden attributes, creation time extraction, symlink handling, and rename behavior.
