# Product Context

## Problem Being Solved
Game launchers need to verify local file integrity and download missing or corrupted files. This server acts as the authoritative source of truth: it knows what files exist, their sizes, and their hashes, so clients can compare local state against the manifest and only download what's needed.

## Why This Approach
- **Stateless manifest**: The manifest is computed from the actual files on disk, so there's no separate database to keep in sync.
- **On-demand recalculation**: When the server operator updates game files, they call `POST /manifest/recalculate` rather than restarting the server.
- **Soft files**: The `?pattern` prefix in `.manifestignore` marks files as optional — clients download them only if they don't already have a local copy. Useful for user-generated content, save files, or large optional assets.
- **Streaming MD5**: Files are hashed with `io.Copy` to avoid loading large game files fully into memory.

## Intended Users
- **Server operators**: Deploy this as a Docker container, point it at a directory of game build artifacts, and expose it to launcher clients.
- **Launcher clients** (not in this repo): Read the manifest, diff against local files, and download from `/files/`.

## User Experience Goals
- Simple deployment — two env vars and done
- Fast manifest serving (manifest is computed once and cached in memory)
- Safe file serving with no path traversal risk
