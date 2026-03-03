# Tech Context

## Language & Runtime
- **Go 1.25** (`go.mod`: `module launcher-host`)
- No external dependencies — stdlib only

## Module Layout
```
launcher-host/
├── internal/
│   ├── cmd/main.go          # Entry point
│   ├── manifest/
│   │   ├── files.go         # HashedFile, Manifest, HashFile, CreateManifest
│   │   ├── files_test.go    # Equivalence tests + benchmarks (1/10/100 MB)
│   │   ├── ignore.go        # FileMode, IgnoreRule, IgnoreRules, ParseIgnoreFile, Match
│   │   └── ignore_test.go   # Ignore rule unit tests + integration test
│   └── server/
│       └── server.go        # Server struct, HTTP handlers, ListenAndServe
├── Dockerfile
└── go.mod
```

## Key Types
| Type | Package | Description |
|---|---|---|
| `HashedFile` | `manifest` | `{path, hash, size, mode}` — one file entry |
| `Manifest` | `manifest` | `{files []*HashedFile}` — full manifest |
| `FileMode` | `manifest` | `""` = required, `"soft"` = optional |
| `IgnoreRule` | `manifest` | `{Pattern string, Soft bool}` |
| `IgnoreRules` | `manifest` | Slice of rules; `Match()` uses last-match-wins |
| `Server` | `server` | `{basePath, currentManifest, mu sync.RWMutex}` |

## Docker
- Multi-stage build: `golang:1.25-alpine` → `alpine:3.19`
- Exposes port 8080
- Binary at `/launcher-server`

## Development Notes
- `.game/` directory is gitignored — used as local test file store
- Tests use `t.TempDir()` for isolated file system state
- `hashFileReadAll` in `files.go` is kept solely for benchmark comparison; not used in production
