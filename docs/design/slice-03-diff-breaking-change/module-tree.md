# module-tree — slice-03-diff-breaking-change

Owns package: `internal/impact/`. Один вход (`ImpactRequest`) = один результат (`BreakingChange | Error`).

## Дерево узлов

```
AssessImpact (head: линейная ROP-труба)                    internal/impact/head.go          io: none
├── ParseImpactBody (ingress-адаптер)                      internal/impact/adapter.go       io: none [BAD_REQUEST]
│     └── NewSpecSource (конструктор exactly-one)          internal/impact/domain.go        io: none
├── SpecLoader.Load ×2 (I/O: path|URL → OpenAPI-док)       internal/shared/specloader/loader.go io: http [SPEC_NOT_FOUND, SPEC_INVALID, ASYNC_IMPACT_NOT_IMPLEMENTED]
├── DiffSpecs (logic: oasdiff old→new → затронутые субъекты) internal/impact/logic.go       io: none
├── AffectedConsumers (logic: субъекты × ConsumersOf)      internal/impact/logic.go          io: none
└── Store (I/O-объект: PrevSpecVersion/ConsumersOf/LastCapturedHash) internal/shared/store io: db
```

## Head-pipe pseudocode

```
AssessImpact(req ImpactRequest, d Deps) -> Result[BreakingChange, Error]:
    | ParseImpactBody(req.Body)  -> ImpactInput            [BAD_REQUEST]
    | d.SpecLoader.Load(from)    -> SpecDoc                [SPEC_NOT_FOUND, SPEC_INVALID, ASYNC_IMPACT_NOT_IMPLEMENTED]
    | d.SpecLoader.Load(to)      -> SpecDoc                [те же]
    | DiffSpecs(from, to)        -> []AffectedSubject      -- oasdiff; METHOD /path
    | d.Store.ConsumersOf(p)     -> []Edge                 -- from_version = PrevSpecVersion
    | AffectedConsumers(...)     -> BreakingChange         -- задетые рёбра; обновить LastCapturedHash(to)
    -> Ok(BreakingChange)
```

## Юнит-тесты (формула N = 1 happy + Σ; head/IO/adapter — компонентными)

| Модуль | Happy | Ветви | Units |
|---|---|---|---|
| `NewSpecSource` | 1 | оба ключа заданы; ни один не задан | 3 |
| `DiffSpecs` | 1 | удалена операция; переименовано поле ответа; ужесточён тип запроса; изменений нет | 5 |
| `AffectedConsumers` | 1 | задетый субъект не потребляется никем; несколько потребителей | 3 |
| **Total** | | | **11** |
