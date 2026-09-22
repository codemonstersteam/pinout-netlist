---
id: ticket-05
type: module
slice: slice-01-ingest-report
blocked_by: [ticket-03, ticket-04]
inputs: [docs/design/slice-01-ingest-report/contracts.md]
outputs: [internal/ingest/{head,adapter,errors,register}.go + маршрутизация POST /reports]
io: none
skills: []
---

# Head-труба IngestReport + ингресс-адаптер (маппинг error.code → 400/409) + wiring в cmd/app.

Компонентные сценарии среза 01 зелёные; @wip снят приёмкой.
