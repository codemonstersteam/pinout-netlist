#!/usr/bin/env bash
# Запуск компонентных тестов template-go-api ВНУТРИ Docker (изоляция; не `go test` с хоста).
set -euo pipefail
cd "$(dirname "$0")/.."
CF=(-f docker-compose.test.yml)
cleanup() { docker compose "${CF[@]}" down -v --remove-orphans >/dev/null 2>&1 || true; }
trap cleanup EXIT
echo "==> building images..."; docker compose "${CF[@]}" build
echo "==> running tests..."; docker compose "${CF[@]}" up --abort-on-container-exit --exit-code-from runner
