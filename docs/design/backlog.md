# backlog — pinout-netlist (E2)

Один тикет на срез. Контракты срезов и пер-срезовые юнит-N дозаполняются по `program-design` при старте E2.

## Тикеты

### T01 — ingest-report (`POST /reports`)
- [ ] `ParseReport` (проверка `schema_version`) + `NewEdge`, `NewVerdictRecord`, `NewSpecVersion` + юнит-тесты.
- [ ] `Store` (I/O): UpsertEdge / AppendVerdict / AppendSpecVersion.
- [ ] head-труба `IngestReport`; ингресс HTTP-handler `POST /reports`.
- [ ] Компонентные тесты: валидный отчёт sync, валидный async, неподдержанный `schema_version`.

### T03 — diff-breaking-change (`POST /impact`)
- [ ] Диф контракта поставщика старая→новая версия (правила из валидаторов / oasdiff).
- [ ] `AffectedConsumers` через `Store.ConsumersOf` + проверка зафиксированного `Subject`.
- [ ] head-труба + ингресс `POST /impact`.
- [ ] Компонентные тесты: безопасное изменение (никого не ломает), breaking-изменение (список затронутых).

### T02 — query-graph (`GET /graph`)
- [ ] `Store.Graph` → проекция; ингресс `GET /graph`.
- [ ] Компонентные тесты: пустой граф, граф с парами.

## Зависимости

- Формат входа готов: [`pinout-openapi/docs/report-format.md`](../../../pinout-openapi/docs/report-format.md).
- Источники отчётов: E0 (`pinout-asyncapi`) и E1 (`pinout-openapi`).
- Собственный API: [`api-specification/openapi.yml`](../../api-specification/openapi.yml).

## Definition of Done (MVP E2 = T01 + T03)

- [ ] Контракты срезов оформлены по `program-design`, граф контрактов сходится.
- [ ] netlist принимает отчёты обоих валидаторов и строит граф пар.
- [ ] Диф двух версий спеки поставщика помечает затронутых потребителей.
- [ ] `gofmt`/`go vet`/`go test`/component-tests зелёные.

## Handoff

- [ ] Оператор аппрувит пакет (после готовности MVP `pinout-openapi`) — @<github-handle>, YYYY-MM-DD
