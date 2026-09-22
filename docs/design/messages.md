# messages — pinout-netlist

Доменная модель графа. Собирается только конструкторами; хранилище за I/O-объектом `Store` (БД не в зависимостях логики). `Result<T, Error>` как в остальных модулях.

> Current as of change greenfield-E2 (lane greenfield). Приведено к канону отчёта 1.1
> (`pinout-openapi/docs/report-format.md`) — разбор долга: `debt/01-report-canon-1.1.md`.

## Доменные типы

```
ServiceId = string                       # имя сервиса (consumer/provider)

SpecVersion {                            # версия контракта сервиса
  Service:  ServiceId
  Ref:      string                        # путь/URL спеки
  Version:  string                        # коммит/тег master
  Kind:     "openapi" | "asyncapi"
}
  NewSpecVersion(raw) → Result<SpecVersion>

Edge {                                   # ребро графа: потребитель → поставщик
  Consumer: ServiceId
  Provider: ServiceId
  Subject:  string                        # операция (sync) / канал (async)
  Interaction: "sync" | "async"
}
  NewEdge(report) → Result<Edge>          # выводится из отчёта валидатора

VerdictRecord {                          # факт проверки во времени
  Edge:        Edge
  Compatible:  bool
  Errors:      []ValidationError          # формат из report-format.md
  ProviderVersion: string
  At:          Timestamp                  # из report.generated_at
}
  NewVerdictRecord(report) → Result<VerdictRecord>

Report                                    # вход: канон pinout-openapi/docs/report-format.md (1.1)
  ParseReport(bytes) → Result<Report>     # проверка schema_version == "1.1" + обязательные поля идентичности

BreakingChange {                         # результат среза 03
  Provider:        ServiceId
  FromVersion, ToVersion: string
  AffectedConsumers: []ServiceId
  Details:         []ValidationError
}
```

## Привязка полей модели к канону 1.1

| Поле модели | Источник в отчёте 1.1 |
|---|---|
| `Edge.Consumer` | `consumer.name` |
| `Edge.Provider` | `provenance.provider` |
| `Edge.Subject` | `errors[].subject` (плюс перечень субъектов из запроса ingest — `subjects[]`) |
| `Edge.Interaction` | `interaction` |
| `VerdictRecord.At` | `generated_at` |
| `VerdictRecord.ProviderVersion` | `provenance.provider_version` |
| `VerdictRecord.Errors` | `errors[]` |
| свежесть provenance | `provenance.captured_hash` |

Идентичность поставщика несёт **только** `provenance{provider, provider_version, captured_hash}`
(канон 1.1); отдельной сущности `provider{}` в модели и спеке нет.

## Следствия варианта A (канон 1.1) — обязательные к исполнению

1. **Агрегатный `compatible`.** Отчёт плоский: один отчёт раскладывается в **N рёбер по
   `errors[].subject`**, а `compatible` в отчёте — агрегатный. Совместимость конкретного ребра
   выводится как «в последнем отчёте **нет ошибки с этим субъектом**», а не читается из отчёта.
   Иначе при реализации родится баг «все рёбра красные, если сломано одно».
2. **Субъекты совместимого отчёта.** При `errors == []` субъектов в отчёте нет (`uncovered_*` —
   только поверхность вне scope). Поэтому запрос ingest — конверт `{report, subjects?[]}`:
   вызывающий (CI потребителя) передаёт свой scope; рёбра = `subjects ∪ error-субъекты ∪ uncovered_*`.
3. **Свежесть provenance.** `provenance.captured_hash` — ключ авто-перепроверки forward по свежести:
   ребро помечается `stale`, когда хеш вердикта не совпадает с последним известным хешем спеки
   поставщика (источник нового хеша — impact-запрос: `to_spec`).
4. **Отчёты exit 2/3 — не вердикты.** Ingest принимает только канонические отчёты 1.1 (схемная
   проверка); конверты без provenance/идентичности отвергаются `BAD_REPORT`.

## Хранилище (I/O, за `Store`)

```
Store.UpsertEdge(Edge) → Result<void>
Store.AppendVerdict(VerdictRecord) → Result<void>
Store.AppendSpecVersion(SpecVersion) → Result<void>
Store.Graph() → Result<[]GraphEdge>          # ребро + последний вердикт (+stale)
Store.ConsumersOf(provider) → Result<[]Edge>
Store.PrevSpecVersion(provider) → Result<SpecVersion>
Store.LastCapturedHash(provider) → Result<string>
```

Движок хранилища — решение реализации (дизайн-пакет не предписывает): MVP — in-memory с
JSON-снапшотом (CGO_ENABLED=0, без внешней БД); интерфейс `Store` изолирует смену движка.

## Семантика breaking-change (срез 03)

Дано: старая и новая версии master-спеки поставщика — **обе передаёт вызывающий** (`from_spec` /
`to_spec`, oneOf путь|URL): netlist спеки целиком не хранит (граница сервиса). Алгоритм:

1. Загрузить обе спеки по источникам (файл/HTTP; io-ошибки → `SPEC_NOT_FOUND`, парсинг → `SPEC_INVALID`).
2. `from_version` — из последнего вердикта поставщика (`Store.PrevSpecVersion`).
3. Диф контракта старая→новая — **oasdiff** (honest reuse); затронутые субъекты `METHOD /path`.
4. Для каждого потребителя из `Store.ConsumersOf(provider)` проверить, задет ли его `Subject` дифом.
5. `BreakingChange{affected_consumers, details}`; заодно обновить `LastCapturedHash` (свежесть).
6. AsyncAPI-спека → `ASYNC_IMPACT_NOT_IMPLEMENTED` (MVP: sync; вынесено в backlog).
