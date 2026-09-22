# Полигон E2E тройки pinout — 7 обязательных сценариев зонтичного TASK.md экосистемы
# (раздел «Полигон»). Отличие от срезовых фич: чёрный ящик — вся тройка
# (реальные бинари pinout-openapi / pinout-asyncapi → отчёт 1.1 → netlist → граф →
# impact), а не один сервис. Запускается отдельно: ./component-tests/scripts/run-polygon.sh
# (GO_TEST_RUN=TestPolygon; обычные компонентные тесты его не гоняют — им не нужны
# бинари валидаторов).
@component @polygon-e2e
Feature: Полигон — тройка pinout работает целиком

  Scenario: P1. sync-пара совместима — зелёное ребро
    Given запущенный сервис netlist
    When бинарь pinout-openapi проверяет пару search-svc и отчёт принят
    Then граф содержит ребро search-svc → catalog с субъектом "GET /items"
    And ребро search-svc → catalog "GET /items" совместимо

  Scenario: P2. sync-пара ломается — красное только по задетому субъекту
    Given запущенный сервис netlist
    When бинарь pinout-openapi проверяет пару checkout-svc и отчёт принят
    Then граф содержит ребро checkout-svc → catalog с субъектом "GET /items"
    And ребро checkout-svc → catalog "GET /items" несовместимо
    And ребро checkout-svc → catalog "POST /orders" совместимо

  Scenario: P3. async-пары — совместимая и ломающаяся
    Given запущенный сервис netlist
    When бинарь pinout-asyncapi проверяет пару /fixtures/validate/good и отчёт принят
    Then граф содержит ребро mq-rest-sync-adapter → wallet-balance с субъектом "WALLET.BALANCE.REQUEST"
    And бинарь pinout-asyncapi проверяет пару /fixtures/validate/incompatible и отчёт принят
    And ребро mq-rest-sync-adapter → wallet-balance "WALLET.FEES.EVENTS" несовместимо
    And ребро mq-rest-sync-adapter → wallet-balance "WALLET.BALANCE.REQUEST" совместимо

  Scenario: P4. impact: breaking v1→v2 задевает checkout-svc и не задевает search-svc
    Given запущенный сервис netlist
    And бинарь pinout-openapi проверяет пару search-svc и отчёт принят
    And бинарь pinout-openapi проверяет пару checkout-svc и отчёт принят
    When клиент отправляет POST /impact с телом из файла /fixtures-ct/polygon/req-poly-breaking.json
    Then ответ 200
    And ответ содержит JSON-поле affected_consumers[0] со значением "checkout-svc"
    And ответ НЕ содержит значение "search-svc"

  Scenario: P5. impact: non-breaking — никого не задевает
    Given запущенный сервис netlist
    And бинарь pinout-openapi проверяет пару checkout-svc и отчёт принят
    When клиент отправляет POST /impact с телом из файла /fixtures-ct/polygon/req-poly-added.json
    Then ответ 200
    And ответ содержит JSON-поле affected_consumers со значением "[]"

  Scenario: P6. свежесть provenance: после impact ребро помечено stale
    Given запущенный сервис netlist
    And бинарь pinout-openapi проверяет пару search-svc и отчёт принят
    And клиент отправляет POST /impact с телом из файла /fixtures-ct/polygon/req-poly-breaking.json
    Then ответ 200
    And ребро search-svc → catalog "GET /items" помечено stale

  Scenario: P7. malformed-входы отвергаются
    Given запущенный сервис netlist
    When клиент отправляет POST /reports с телом из файла /fixtures-ct/polygon/report-1.0.json
    Then ответ 409
    And клиент отправляет POST /reports с телом из файла /fixtures-ct/polygon/report-exit3-no-provenance.json
    Then ответ 400
    And клиент отправляет POST /reports с телом из файла /fixtures-ct/polygon/report-broken.json
    Then ответ 400
