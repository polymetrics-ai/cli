# Polymetrics Go CLI Monolith

This project is a Go-only rewrite of Polymetrics as a local-first CLI monolith.

The first implementation target is a dependency-free MVP that proves the product architecture:

- terminal project initialization
- embedded help and man-style documentation
- encrypted local credentials
- connector registry
- sample/file sources
- local JSONL warehouse destination
- ETL run lifecycle
- reverse ETL plan, preview, approval, and execution
- stable JSON output for agents

