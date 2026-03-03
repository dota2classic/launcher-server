# Active Context

## Current State
Early-stage project. Core functionality is complete and working:
- Manifest creation with ignore rules
- HTTP server with all three endpoints
- Docker build

## Recent Work (from git log)
- `Initial commit` — base structure
- `main` — server + manifest packages implemented
- `wip` — latest commit; exact scope unclear

## Untracked Files
- `internal/manifest/ignore_test.go` — new test file not yet committed

## Open Questions / Possible Next Steps
- File watching: auto-recalculate manifest when files change (currently manual via POST)
- Authentication: no auth on any endpoint — anyone can recalculate or download
- HTTPS support
- Metrics/health endpoint
- Parallel hashing for large file sets (currently sequential WalkDir)

## Active Design Decisions
- Manifest is fully recomputed on recalculate (no incremental/delta)
- No persistent storage — manifest lives in memory only
- No client tracking — server is stateless per-request
