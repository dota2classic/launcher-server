#!/usr/bin/env bash
set -euo pipefail

cd /opt/launcher
docker compose kill -s SIGHUP launcher
echo "Manifest recalculation triggered."
