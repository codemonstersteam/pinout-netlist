---
id: ticket-03
type: module
slice: slice-01-ingest-report
blocked_by: [ticket-02]
inputs: [docs/design/slice-01-ingest-report/contracts.md, api-specification/openapi.yaml]
outputs: [internal/ingest/{domain,logic}.go + юниты]
io: none
skills: []
---

# Домен и логика среза 01: NewReport (канон 1.1: schema_version/идентичность/инвариант), SubjectsOf, NewEdge/NewVerdictRecord/NewSpecVersion — valid by construction.

Юниты по формуле (15) зелёные; validate-constructors зелёный.
