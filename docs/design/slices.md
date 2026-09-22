# slices — pinout-netlist

Один внешний вход = один срез = один Request. Контракты срезов дозаполнены по `program-design` при старте E2 (2026-09-22, greenfield-прогон).

| Внешний вход | ID | Срез | Назначение | Owns package |
|---|---|---|---|---|
| `POST /reports` | 01 | ingest-report | Принять отчёт валидатора (канон 1.1), обновить граф и историю версий | `internal/ingest/` |
| `GET /graph` | 02 | query-graph | Отдать текущий граф пар consumer↔provider с последним вердиктом | `internal/graph/` |
| `POST /impact` (новая версия спеки поставщика) | 03 | diff-breaking-change | Диф контракта old→new (oasdiff) → список затронутых потребителей | `internal/impact/` |

**Порядок:** 01 → 02 → 03. 01 наполняет хранилище; 02 читает; 03 опирается на накопленную историю версий из 01.

MVP E2 = срезы 01 + 03 (приём отчётов и детект breaking-change). 02 — следом (тривиальная проекция Store).

Общее для срезов (лениво в shared, используют ≥2 среза): `internal/shared/config` (конфиг сервиса), `internal/shared/store` (порт + JSON-снапшот реализация), `internal/shared/specloader` (загрузчик спек path|URL).
