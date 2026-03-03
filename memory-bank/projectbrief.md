# Project Brief

## Project Name
`launcher-host` — Game Launcher Host Server

## Core Purpose
An HTTP server that scans a directory of game files, computes MD5 hashes for each file, and exposes a manifest + file-serving API so a game launcher client can verify integrity and download files.

## Key Requirements
- Scan a base directory and produce a JSON manifest of all files with path, MD5 hash, size, and mode
- Serve files over HTTP with directory traversal protection
- Support `.manifestignore` to exclude files or mark them as "soft" (optional: download only if missing locally)
- Allow on-demand manifest recalculation without restarting the server
- Be deployable as a Docker container

## Environment Variables
| Variable | Required | Default | Description |
|---|---|---|---|
| `LAUNCHER_FILES_PATH` | Yes | — | Base directory to scan |
| `LAUNCHER_ADDR` | No | `:8080` | Listen address |

## HTTP API
| Method | Path | Description |
|---|---|---|
| `GET` | `/manifest` | Returns full manifest as JSON |
| `POST` | `/manifest/recalculate` | Triggers manifest recalculation |
| `GET` | `/files/{path}` | Downloads a specific file |
