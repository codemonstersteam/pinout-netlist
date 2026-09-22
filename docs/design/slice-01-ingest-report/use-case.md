# use-case — slice-01-ingest-report

**UC-1: Принять отчёт валидатора** · Level: user-goal · Trigger: CI потребителя (или оператор) отправляет `POST /reports` с конвертом `{report, subjects?[]}`.

## Main success scenario (MSS)

1. Оператор CI отправляет отчёт валидатора (канон 1.1: `pinout-openapi` или `pinout-asyncapi`) и опционально перечень субъектов scope.
2. netlist проверяет `schema_version == "1.1"` и обязательные поля идентичности (`validator`, `interaction`, `consumer.name`, `provenance`).
3. netlist выводит субъекты: `subjects[] ∪ errors[].subject ∪ uncovered_*`.
4. Для каждого субъекта — `Edge{consumer, provider, subject, interaction}` идемпотентно upsert-ится; на каждое ребро appended `VerdictRecord{compatible_ребра, errors_субъекта, provider_version, generated_at}`; `SpecVersion` поставщика обновляется (`captured_hash` — ключ свежести).
5. netlist отвечает `202 {edges_accepted}`.

## Extensions (1 Extension = 1 error.code = 1 компонентный сценарий)

| # | Ветвь | error.code | HTTP |
|---|---|---|---|
| 1a | Тело — не канонический отчёт (битый JSON, нет обязательных полей идентичности/provenance, exit-3-конверт) | `BAD_REPORT` | 400 |
| 1b | `schema_version ≠ "1.1"` | `UNSUPPORTED_SCHEMA_VERSION` | 409 |

Совместимость ребра = «нет ошибки с этим субъектом» (агрегатный `compatible` разложен по субъектам — messages.md, следствие 1).
