# Realized 1:1 from docs/design/slice-02-query-graph/contracts.md
# §Component scenarios. N = 1 (happy) + 0 adapter branches = 1.
@component @slice-02-query-graph
Feature: Query the consumer↔provider graph

  Scenario: граф отдаёт рёбра после ingest
    Given запущенный сервис netlist
    When клиент отправляет POST /reports с телом из файла /fixtures/ingest/valid-sync.json
    Then ответ 202
    And клиент отправляет GET /graph
    Then ответ 200
    And граф содержит ребро search-svc → catalog с субъектом "GET /items"
    And ответ содержит JSON-поле edges[0].compatible со значением "true"
