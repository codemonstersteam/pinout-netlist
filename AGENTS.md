# AGENTS.md — pinout-netlist

Точка входа для агента/разработчика. **Порядок чтения:** этот файл → [`CLAUDE.md`](./CLAUDE.md) → пакет проектирования [`docs/design/`](./docs/design/) → собственный контракт [`api-specification/openapi.yml`](./api-specification/openapi.yml).

## Что это

Координатор графа связей сервисов: принимает отчёты валидаторов в [общем формате](https://github.com/codemonstersteam/pinout-openapi/blob/main/docs/report-format.md), строит граф `consumer↔provider`, детектит breaking-change во времени. Концепт: [`../pinout/README.md`](../pinout/README.md). Эпик **E2** в [`../pinout/backlog.md`](../pinout/backlog.md). Это сетевой сервис (в отличие от CLI `pinout-openapi`).

## Разработка: харнес izi

Разработку ведёт **харнес `izi`** ([rationaldev-ai-sdlc-skills](https://github.com/codemonstersteam/rationaldev-ai-sdlc-skills/)) — его скиллы. Правила — в [`CLAUDE.md`](./CLAUDE.md).

## Resume here — откуда продолжить

**Состояние:** каркас + пакет проектирования верхнего уровня готовы; **код не начат**.

**Следующий шаг:** стартует **после** MVP `pinout-openapi` (зависимость по бэклогу). Перед реализацией — дозаполнить пер-срезовые контракты по `program-design`.

- Срезы: `ingest-report` (01), `query-graph` (02), `diff-breaking-change` (03) — [`docs/design/slices.md`](./docs/design/slices.md). MVP E2 = 01 + 03.
- Доменная модель графа: [`docs/design/messages.md`](./docs/design/messages.md). Тикеты/DoD: [`docs/design/backlog.md`](./docs/design/backlog.md).
- Формат входа готов (канон в `pinout-openapi`).

**Gate:** реализация — после handoff-аппрува оператора в [`docs/design/backlog.md`](./docs/design/backlog.md).
