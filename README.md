# pinout-netlist

> Part of the **pinout** contract-testing platform — концепт: [`pinout`](https://github.com/codemonstersteam/pinout) ([локально](../pinout/README.md)). Название — отсылка к *netlist* в PCB-дизайне: формальное описание всех соединений схемы.

Координатор графа связей сервисов: хранит `consumer → provider` с историей вердиктов и отвечает, **кто сломается, если поставщик изменит контракт** (breaking-change во времени).

## Can / Cannot

**Can:**

- Принимать отчёты обоих валидаторов (канон 1.1) и строить граф рёбер по субъектам.
- Отвечать на impact-запрос: диф двух версий спеки поставщика (`oasdiff`) → затронутые живущие потребители.
- Хранить историю версий и свежесть `provenance.captured_hash` (пометка `stale`).
- Переживать рестарт (JSON-снапшот хранилища).

**Cannot:**

- Валидировать пару consumer↔provider (это [`pinout-openapi`](https://github.com/codemonstersteam/pinout-openapi) / [`pinout-asyncapi`](https://github.com/codemonstersteam/pinout-asyncapi)).
- Хранить спеки целиком (хранит рёбра, вердикты, версии и хеши).
- Считать impact для AsyncAPI-поставщиков (MVP: sync; backlog E2).
- Генерировать consumed-contract или стабы (это E-harness).

## Stack

| Component | Technology |
|---|---|
| Language / runtime | Go (CGO_ENABLED=0, stdlib `net/http`) |
| Contract | [`api-specification/openapi.yaml`](api-specification/openapi.yaml) (`x-frozen: 2026-09-22`) |
| Input format | Отчёт канона 1.1 — [report-format](https://github.com/codemonstersteam/pinout-openapi/blob/main/docs/report-format.md) |
| Spec diff | `oasdiff` + `kin-openapi` (honest reuse) |
| Storage | in-memory + JSON-снапшот (за портом `Store`) |

## API

| Endpoint | Method | Action |
|---|---|---|
| `/reports` | POST | Ingest отчёта валидатора `{report, subjects?}` → `202 {edges_accepted}` \| `400` \| `409` |
| `/graph` | GET | Текущий граф рёбер с последним вердиктом → `200 {edges}` |
| `/impact` | POST | `{provider, to_version, from_spec, to_spec}` → `200 BreakingChange` \| ошибки |
| `/health` | GET | Liveness → `200` |

## Как работает (pipe — где работает и где ломается)

```text
POST /reports
| Parse body {report, subjects?}                          [BAD_REPORT → 400]
| Канон 1.1: schema_version/идентичность/provenance      [UNSUPPORTED_SCHEMA_VERSION → 409]
| Субъекты = subjects ∪ errors[].subject ∪ uncovered_*    → N рёбер (compatible ребра = «нет ошибки с этим субъектом»)
| Store: upsert рёбер + append вердикты/версии            → 202 {edges_accepted}

POST /impact
| Parse body {provider, to_version, from_spec, to_spec}   [BAD_REQUEST → 400]
| Загрузить обе спеки (path|URL)                          [SPEC_NOT_FOUND → 404, SPEC_INVALID → 422, AsyncAPI → 501]
| oasdiff old→new → затронутые субъекты METHOD /path
| Store.ConsumersOf(provider) → задетые рёбра             → 200 {affected_consumers, details}
```

## Карта режимов отказа (failure map)

| HTTP | error.code | Смысл |
|---|---|---|
| 400 | `BAD_REPORT` | тело — не канонический отчёт 1.1 (нет идентичности/provenance, exit-3-конверт, битый JSON) |
| 409 | `UNSUPPORTED_SCHEMA_VERSION` | `schema_version ≠ "1.1"` |
| 400 | `BAD_REQUEST` | impact-запрос не соответствует схеме (источник не exactly-one) |
| 404 | `SPEC_NOT_FOUND` | источник спеки недоступен (файл отсутствует / HTTP не-2xx) |
| 422 | `SPEC_INVALID` | спека не парсится как OpenAPI 3.x |
| 501 | `ASYNC_IMPACT_NOT_IMPLEMENTED` | async-поставщик: вынесено в backlog E2 |
| 501 | `NOT_IMPLEMENTED` | scaffold-заглушка нереализованного среза |

Правило: всё непроверенное или деградировавшее видно в ответе (`error.code` + статус) — никогда не маскируется под успех.

## Build & run

```bash
go build -o pinout-netlist ./cmd/app
./pinout-netlist                       # :8080 (config.yaml: listen_addr, store_file)

curl -s localhost:8080/health
curl -s -X POST localhost:8080/reports -d @report.json       # отчёт валидатора 1.1
curl -s localhost:8080/graph
```

Поведение снаружи доказывается [`component-tests/`](component-tests/) (в Docker Compose; полигон E2E — с реальными бинарями обоих валидаторов).

## Learn more (retrievability ladder)

1. **This README** — что это и как запустить.
2. [`component-tests/`](component-tests/) — поведение снаружи (чёрный ящик + полигон).
3. `docs/design/` — [intent](docs/design/intent.md), [messages](docs/design/messages.md) (модель + привязка к канону 1.1), [slices](docs/design/slices.md), [backlog](docs/design/backlog.md); дизайн-пакеты срезов: `docs/design/slice-0{1,2,3}-*/` (use-case, module-tree, contracts, c4, tickets).
4. Platform architecture: [`pinout/docs/CONCEPT.md`](https://github.com/codemonstersteam/pinout/blob/main/docs/CONCEPT.md); канон отчёта: [`pinout-openapi/docs/report-format.md`](https://github.com/codemonstersteam/pinout-openapi/blob/main/docs/report-format.md) (владелец — эпик E1).
