---
id: ticket-04
type: module
slice: slice-01-ingest-report
blocked_by: [ticket-03]
inputs: [docs/design/messages.md]
outputs: [internal/shared/store/store.go + снапшот + юниты]
io: db
skills: [db-io, db-schema]
---

# I/O-объект Store: порт + in-memory реализация с JSON-снапшотом; UpsertEdge/AppendVerdict/AppendSpecVersion/Graph/ConsumersOf/PrevSpecVersion/LastCapturedHash.

Store юниты зелёные; снапшот переживает рестарт.
