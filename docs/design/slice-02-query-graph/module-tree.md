# module-tree — slice-02-query-graph

Owns package: `internal/graph/`. Один вход (`GraphRequest{}`) = один результат (`Graph`).

## Дерево узлов

```
QueryGraph (head: чтение хранилища → проекция)   internal/graph/head.go           io: none
└── Store.Graph() (I/O-объект)                    internal/shared/store/store.go   io: db
```

## Head-pipe pseudocode

```
QueryGraph(req GraphRequest, d Deps) -> Result[Graph, Error]:
    | d.Store.Graph()  -> []GraphEdge   -- ребро + последний вердикт (+stale); пустое хранилище → []
    -> Ok(Graph{Edges})
```

## Юнит-тесты (формула N = 1 happy + Σ; head/IO — компонентными)

| Модуль | Happy | Ветви | Units |
|---|---|---|---|
| `projectEdge` (проекция записи → GraphEdge DTO) | 1 | stale-пометка проставляется | 2 |
| **Total** | | | **2** |
