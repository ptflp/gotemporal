#!/usr/bin/env bash
set -euo pipefail

cd /app

if [ ! -f go.mod ]; then
  echo "go.mod not found in /app. Did you mount the project directory?" >&2
  exit 1
fi

# Keep module/cache dirs in sync for faster rebuilds
export GOMODCACHE=${GOMODCACHE:-/go/pkg/mod}
export GOCACHE=${GOCACHE:-/root/.cache/go-build}

echo "[app] downloading Go modules..."
go mod download

echo "[app] building binary..."
go build -buildvcs=false -o /tmp/app ./cmd/app

echo "[app] starting service..."
exec /tmp/app

