# Progress

## Done
- [x] `manifest` package: `HashedFile`, `Manifest` types with JSON tags
- [x] Streaming MD5 hashing (`HashFile`)
- [x] `.manifestignore` parsing (comments, soft `?` prefix, empty lines)
- [x] Pattern matching: exact, wildcard, directory (`dir/`), double wildcard (`**`)
- [x] `CreateManifest`: walks dir, applies ignore rules, returns manifest
- [x] `Server` struct with `sync.RWMutex` for thread-safe manifest access
- [x] `GET /manifest` endpoint
- [x] `POST /manifest/recalculate` endpoint
- [x] `GET /files/{path}` with directory traversal protection
- [x] Multi-stage Dockerfile
- [x] Tests: hash equivalence, benchmarks (1/10/100MB), ignore rule unit tests, manifest integration test

## In Progress / Pending
- [ ] Commit `ignore_test.go` (currently untracked)
- [ ] Decide on file watcher for auto-recalculation
- [ ] Consider auth layer (API key / token) for write endpoints
- [ ] Consider parallel manifest computation for large directories

## Known Issues / Limitations
- `ModeIgnored` and `ModeRequired` both map to `""` — `ModeIgnored` is purely internal, but the const naming could be confusing
- No health check endpoint (useful for Docker/k8s readiness probes)
- Sequential file hashing may be slow for directories with thousands of files
