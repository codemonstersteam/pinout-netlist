# messages — pinout-netlist

Доменная модель графа. Собирается только конструкторами; хранилище за I/O-объектом `Store` (БД не в зависимостях логики). `Result<T, Error>` как в остальных модулях.

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

Report                                    # вход: канон pinout-openapi/docs/report-format.md
  ParseReport(bytes) → Result<Report>     # проверка schema_version

BreakingChange {                         # результат среза 03
  Provider:        ServiceId
  FromVersion, ToVersion: string
  AffectedConsumers: []ServiceId
  Details:         []ValidationError
}
```

## Хранилище (I/O, за `Store`)

```
Store.UpsertEdge(Edge) → Result<void>
Store.AppendVerdict(VerdictRecord) → Result<void>
Store.AppendSpecVersion(SpecVersion) → Result<void>
Store.Graph() → Result<[]Edge>
Store.ConsumersOf(provider) → Result<[]Edge>
Store.PrevSpecVersion(provider) → Result<SpecVersion>
```

## Семантика breaking-change (срез 03)

Дано: новая версия master-спеки поставщика. Алгоритм:
1. Взять предыдущую `SpecVersion` поставщика (`Store.PrevSpecVersion`).
2. Диф контракта старая→новая (переиспользуем правила сравнения схем из валидаторов / oasdiff).
3. Для каждого потребителя из `Store.ConsumersOf(provider)` проверить, ломает ли диф его зафиксированный `Subject`.
4. Вернуть `BreakingChange` со списком затронутых потребителей.
