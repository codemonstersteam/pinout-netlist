# contracts — slice-02-query-graph

## Карточки узлов

### QueryGraph (head, io: none)

- **Signature:** `QueryGraph(d Deps) -> Result[Graph, Error]`
- **Input (data):** — (GET без параметров)
- **Dependencies:** `Store` (порт)
- **io:** none (чтение через I/O-объект Store)
- **What it does:** проекция хранилища в `Graph{edges: [GraphEdge]}`; последнее состояние ребра (compatible, last_verdict_at, provider_version, stale?).
- **Antecedent:** сервис запущен.
- **Consequent:** `200 Graph` всегда (пустой граф — `edges: []`); ветвей отказа нет.

### Store.Graph (io: db)

- **Signature:** см. `docs/design/messages.md`; возвращает рёбра с последним вердиктом каждого.

## Component scenarios (DESIGN half)

Формула `N = 1 (happy) + Σ (0 adapter branches) = 1`.

| # | Сценарий | Статус | Утверждение | тег |
|---|---|---|---|---|
| 1 | граф отдаёт рёбра после ingest | 200 | `edges` непуст; поля ребра идентичны последнему вердикту | accepted |

```gherkin
@component @slice-02-query-graph
Feature: Query the consumer↔provider graph

  Scenario: граф отдаёт рёбра после ingest
    Given запущенный сервис netlist и принятый отчёт
    When клиент отправляет GET /graph
    Then ответ 200 и edges содержит ребро с последним вердиктом
```
