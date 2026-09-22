# Realized 1:1 from docs/design/slice-01-ingest-report/contracts.md
# §Component scenarios. N = 1 (happy) + 2 (adapter branches: BAD_REPORT,
# UNSUPPORTED_SCHEMA_VERSION) = 3. Добавление/удаление сценария — акт дизайна.
@component @slice-01-ingest-report
Feature: Ingest validator reports (canon 1.1)

  Scenario: валидный sync-отчёт принят
    Given запущенный сервис netlist
    When клиент отправляет POST /reports с телом из файла /fixtures/ingest/valid-sync.json
    Then ответ 202
    And ответ содержит непустое JSON-поле edges_accepted

  Scenario: неканонический отчёт отвергнут
    Given запущенный сервис netlist
    When клиент отправляет POST /reports с телом из файла /fixtures/ingest/bad-report.json
    Then ответ 400
    And ответ содержит JSON-поле error.code со значением "BAD_REPORT"

  Scenario: неподдержанная версия схемы отвергнута
    Given запущенный сервис netlist
    When клиент отправляет POST /reports с телом из файла /fixtures/ingest/schema-1.0.json
    Then ответ 409
    And ответ содержит JSON-поле error.code со значением "UNSUPPORTED_SCHEMA_VERSION"
