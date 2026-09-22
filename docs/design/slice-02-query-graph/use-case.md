# use-case — slice-02-query-graph

**UC-2: Отдать граф связей** · Level: user-goal · Trigger: оператор/CI отправляет `GET /graph`.

## Main success scenario (MSS)

1. Оператор запрашивает текущий граф.
2. netlist проецирует хранилище: `GraphEdge{consumer, provider, subject, interaction, compatible, last_verdict_at, provider_version, stale?}` — последнее состояние каждого ребра.
3. Ответ `200 {edges}` (пустое хранилище → `edges: []` — не ошибка).

## Extensions

Нет собственных ветвей отказа (чтение без внешних зависимостей; пустой граф — валидный ответ). Формула компонентных сценариев: `N = 1` (happy).
