# module-tree — slice-01-ingest-report

Owns package: `internal/ingest/`. Один внешний вход = один `Request` (`IngestRequest{Body []byte}`) = один `Result` (`IngestAccepted | Error`).

## Дерево узлов

```
IngestReport (head: линейная ROP-труба)                    internal/ingest/head.go        io: none
├── ParseIngestBody (ingress-адаптер: JSON → IngestInput)  internal/ingest/adapter.go    io: none [BAD_REPORT]
│     └── ParseReport (конструктор канона 1.1)             internal/ingest/domain.go     io: none [UNSUPPORTED_SCHEMA_VERSION, BAD_REPORT]
├── SubjectsOf (logic: subjects ∪ error-субъекты ∪ uncovered) internal/ingest/logic.go   io: none
├── NewEdge / NewVerdictRecord / NewSpecVersion (конструкторы) internal/ingest/domain.go io: none
└── Store (I/O-объект: граф+история, JSON-снапшот)         internal/shared/store/store.go io: db
      ├── UpsertEdge / AppendVerdict / AppendSpecVersion / LastCapturedHash
```

HTTP-маппинг (`error.code` → статус) — только в ингресс-адаптере `cmd/app` (роут) + `adapter.go`.

## Head-pipe pseudocode

```
IngestReport(req IngestRequest, d Deps) -> Result[IngestAccepted, Error]:
    | ParseIngestBody(req.Body)      -> IngestInput         [BAD_REPORT / UNSUPPORTED_SCHEMA_VERSION]
    | SubjectsOf(input)              -> []subject           -- subjects ∪ error-субъекты ∪ uncovered_*
    | NewEdge/NewVerdictRecord/...   -> обновления          -- конструкторы, valid by construction
    | d.Store.Apply(обновления)      -> edges_accepted      [ошибка хранилища → 500 INTERNAL]
    -> Ok(IngestAccepted{EdgesAccepted})
```

## Юнит-тесты (формула N = 1 happy + Σ ветвей конструкторов/pure logic; head/IO/adapter — компонентными)

| Модуль | Happy | Ветви | Units |
|---|---|---|---|
| `ParseReport` | 1 | schema_version ≠ 1.1; нет consumer.name; нет provenance; errors != [] при compatible; битый JSON | 6 |
| `SubjectsOf` | 1 | только subjects; только error-субъекты; только uncovered; дубликат схлопывается | 5 |
| `NewEdge` | 1 | пустой субъект | 2 |
| `NewVerdictRecord` | 1 | совместимость ребра при ошибке чужого субъекта | 2 |
| **Total** | | | **15** |
