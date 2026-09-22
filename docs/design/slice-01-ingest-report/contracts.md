# contracts — slice-01-ingest-report

## Карточки узлов

### ParseIngestBody (adapter, io: none)

- **Signature:** `ParseIngestBody(body []byte) -> Result[IngestInput, Error]`
- **Input (data):** сырые байты тела `POST /reports`.
- **Dependencies:** —
- **io:** none (json.Unmarshal; сеть отсутствует)
- **What it does:** JSON-декод конверта `{report, subjects?}`; делегирует проверку канона `ParseReport`.
- **Antecedent:** HTTP-тело запроса.
- **Consequent:** `IngestInput` или `ErrBadReport` (400).

### ParseReport (constructor, io: none)

- **Signature:** `NewReport(raw RawReport) -> Result[Report, Error]`
- **Input (data):** декодированный отчёт.
- **Dependencies:** —
- **io:** none
- **What it does:** канон 1.1: `schema_version == "1.1"` (иначе `UNSUPPORTED_SCHEMA_VERSION`), непустые `validator/interaction/consumer.name/provenance`, инвариант `compatible ⇔ errors == []`.
- **Consequent:** валидный `Report` (valid by construction).

### SubjectsOf (pure logic, io: none)

- **Signature:** `SubjectsOf(r Report, subjects []string) -> []string` — дедуплицированное объединение `subjects ∪ errors[].subject ∪ uncovered_*`.

### Store (I/O-объект, io: db)

- **Signature:** `Apply(u StoreUpdate) -> Result[int, Error]`; см. полный порт в `docs/design/messages.md`.
- **io:** db (in-memory + JSON-снапшот; БД не в зависимостях логики)
- **What it does:** идемпотентный upsert рёбер + append вердиктов/версий; обновление `LastCapturedHash`.

## Component scenarios (DESIGN half)

Формула `N = 1 (happy) + Σ distinguishable adapter branches (2) = 3`. Гейт: `#сценариев == #ветвей == #error.codes` (BAD_REPORT, UNSUPPORTED_SCHEMA_VERSION).

| # | Сценарий | exit-статус | Утверждение | тег |
|---|---|---|---|---|
| 1 | валидный sync-отчёт → ingest | 202 | `edges_accepted ≥ 1`; ребро в `/graph` с `compatible` и `last_verdict_at` | accepted |
| 2 | неканонический отчёт (нет provenance / exit-3-конверт / битый JSON) | 400 | `error.code == "BAD_REPORT"` | accepted |
| 3 | `schema_version: "1.0"` | 409 | `error.code == "UNSUPPORTED_SCHEMA_VERSION"` | accepted |

```gherkin
@component @slice-01-ingest-report
Feature: Ingest validator reports (canon 1.1)

  Scenario: валидный sync-отчёт принят
    Given запущенный сервис netlist
    When клиент отправляет POST /reports с валидным sync-отчётом 1.1 и субъектами
    Then ответ 202 и edges_accepted >= 1
    And GET /graph содержит ребро с consumer, provider, subject и compatible

  Scenario: неканонический отчёт отвергнут
    When клиент отправляет POST /reports с отчётом без provenance
    Then ответ 400 и error.code == "BAD_REPORT"

  Scenario: неподдержанная версия схемы отвергнута
    When клиент отправляет POST /reports с schema_version "1.0"
    Then ответ 409 и error.code == "UNSUPPORTED_SCHEMA_VERSION"
```
