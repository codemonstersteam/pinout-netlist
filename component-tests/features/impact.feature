# Realized 1:1 from docs/design/slice-03-diff-breaking-change/contracts.md
# §Component scenarios. N = 2 (happy-class: breaking + non-breaking) + 4 adapter
# branches (BAD_REQUEST, SPEC_NOT_FOUND, SPEC_INVALID, ASYNC_IMPACT_NOT_IMPLEMENTED) = 6.
@component @slice-03-diff-breaking-change
Feature: Assess the impact of a new provider spec version

  Scenario: breaking-изменение задевает потребителя
    Given запущенный сервис netlist
    And клиент отправляет POST /reports с телом из файла /fixtures/impact/checkout-report.json
    Then ответ 202
    When клиент отправляет POST /impact с телом из файла /fixtures/impact/req-breaking.json
    Then ответ 200
    And ответ содержит JSON-поле affected_consumers[0] со значением "checkout-svc"
    And ответ содержит JSON-поле details[0].subject со значением "POST /orders"

  Scenario: non-breaking изменение никого не задевает
    Given запущенный сервис netlist
    And клиент отправляет POST /reports с телом из файла /fixtures/impact/checkout-report.json
    Then ответ 202
    When клиент отправляет POST /impact с телом из файла /fixtures/impact/req-added.json
    Then ответ 200
    And ответ содержит JSON-поле affected_consumers со значением "[]"

  Scenario: запрос без источников спек
    Given запущенный сервис netlist
    When клиент отправляет POST /impact с телом из файла /fixtures/impact/req-empty.json
    Then ответ 400
    And ответ содержит JSON-поле error.code со значением "BAD_REQUEST"

  Scenario: файл спеки отсутствует
    Given запущенный сервис netlist
    When клиент отправляет POST /impact с телом из файла /fixtures/impact/req-missing-spec.json
    Then ответ 404
    And ответ содержит JSON-поле error.code со значением "SPEC_NOT_FOUND"

  Scenario: спека невалидна
    Given запущенный сервис netlist
    When клиент отправляет POST /impact с телом из файла /fixtures/impact/req-broken-spec.json
    Then ответ 422
    And ответ содержит JSON-поле error.code со значением "SPEC_INVALID"

  Scenario: async-поставщик не поддержан
    Given запущенный сервис netlist
    When клиент отправляет POST /impact с телом из файла /fixtures/impact/req-async-spec.json
    Then ответ 501
    And ответ содержит JSON-поле error.code со значением "ASYNC_IMPACT_NOT_IMPLEMENTED"
