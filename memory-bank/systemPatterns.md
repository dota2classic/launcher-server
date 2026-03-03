# System Patterns

## Architecture
Single-binary HTTP server. No external services. All state is in-process.

```
main.go
  └── server.New(basePath)
        ├── manifest.CreateManifest(basePath)   ← runs at startup
        │     ├── ParseIgnoreFile               ← reads .manifestignore
        │     └── filepath.WalkDir + HashFile   ← streams MD5 per file
        └── srv.ListenAndServe(addr)
              ├── GET  /manifest                ← return cached manifest (RLock)
              ├── POST /manifest/recalculate    ← recompute manifest (Lock)
              └── GET  /files/{path}            ← serve file from disk
```

## Concurrency Pattern
`Server` uses `sync.RWMutex`:
- `RLock` for reads (`/manifest`, post-recalculate JSON response)
- `Lock` for writes (`recalculateManifest`)

## Ignore File Semantics (`.manifestignore`)
- Comments: lines starting with `#`
- Ignored: plain pattern line → file excluded from manifest
- Soft: `?pattern` → file included in manifest with `mode:"soft"`
- Last-match-wins (like `.gitignore`)
- Pattern types supported:
  - Exact: `config.ini`
  - Wildcard: `*.log`, `temp/*`
  - Directory: `dir/` → matches all files under `dir/`
  - Double wildcard: `**/*.txt` → any depth
- `.manifestignore` itself is always excluded from the manifest

## Security: Directory Traversal Prevention
`handleFiles` applies two guards:
1. `filepath.Clean` + check for `..` prefix or absolute path on the relative component
2. `isSubPath(basePath, fullPath)` using `filepath.Rel` to verify the resolved path is still under basePath

## Path Normalization
All relative paths in the manifest use forward slashes (`filepath.ToSlash`) for cross-platform client compatibility.

## File Hashing
- Algorithm: MD5 (sufficient for integrity checking; not a security hash)
- Method: streaming via `io.Copy(md5.New(), file)` — avoids loading large files into memory
- `hashFileReadAll` exists only as a benchmark baseline; not used in production path
