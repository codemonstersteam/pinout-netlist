#!/usr/bin/env bash
# Полигон E2E тройки pinout ВНУТРИ Docker (netlist + реальные бинари обоих
# валидаторов из братских репо воркспейса). Не `go test` с хоста.
# Двухфазный подъём (как у валидаторов): стейджеры tool-* — one-shot, поэтому
# `up --abort-on-container-exit` убил бы всё при их штатном выходе; раннер
# стартуется отдельным `run --rm` после здоровья сервиса.
set -euo pipefail
cd "$(dirname "$0")/.."
CF=(-f docker-compose.polygon.yml)
cleanup() { docker compose "${CF[@]}" down -v --remove-orphans >/dev/null 2>&1 || true; }
trap cleanup EXIT
echo "==> building images (включая бинари валидаторов из ../pinout-{openapi,asyncapi})..."
docker compose "${CF[@]}" build
echo "==> starting service + tool stagers (detached)..."
docker compose "${CF[@]}" up -d service tool-openapi tool-asyncapi
echo "==> running polygon runner..."
docker compose "${CF[@]}" run --rm runner
