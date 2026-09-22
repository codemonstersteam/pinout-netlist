# contracts — slice-03-diff-breaking-change

## Карточки узлов

### ParseImpactBody (adapter, io: none)

- **Signature:** `ParseImpactBody(body []byte) -> Result[ImpactInput, Error]`
- **io:** none
- **What it does:** JSON-декод `{provider, to_version, from_spec, to_spec}`; `NewSpecSource` проверяет exactly-one (путь|URL) на каждый источник.
- **Consequent:** `ImpactInput` или `ErrBadRequest` (400).

### SpecLoader.Load (I/O, io: http)

- **Signature:** `Load(src SpecSource) -> Result[SpecDoc, Error]`
- **Dependencies:** автономный объект (инкапсулирует `*http.Client`, таймаут, Bearer).
- **io:** http (исходящий Client; файловый путь — fs)
- **What it does:** загрузка спеки по path|URL; парсинг OpenAPI 3.x (kin-openapi, honest reuse — как у `pinout-openapi`); AsyncAPI-документ → `ErrAsyncNotImplemented`.
- **Consequent:** `SpecDoc`; `ErrSpecNotFound` (404) / `ErrSpecInvalid` (422) / `ErrAsyncImpact` (501).

### DiffSpecs (pure logic, io: none)

- **Signature:** `DiffSpecs(from, to SpecDoc) -> []AffectedSubject`
- **What it does:** oasdiff-диф (`honest reuse`): удалённые операции, удалённые/переименованные поля ответа, ужесточение типов запроса → субъекты `METHOD /path` с кодом изменения.
- **Consequent:** значение (может быть пустым = non-breaking).

### AffectedConsumers (pure logic, io: none)

- **Signature:** `AffectedConsumers(subjects []AffectedSubject, edges []Edge) -> []string` — живущие потребители, чьи `Subject` задеты.

## Component scenarios (DESIGN half)

Формула `N = 2 (happy-class: breaking + non-breaking) + Σ distinguishable adapter branches (4) = 6`.

| # | Сценарий | Статус | Утверждение | тег |
|---|---|---|---|---|
| 1 | breaking: удалена операция, потребляемая живущим консьюмером | 200 | `affected_consumers` содержит его; `details[].subject` = операция | accepted |
| 2 | non-breaking: добавлена новая операция | 200 | `affected_consumers == []` | accepted |
| 3 | запрос без источников спек | 400 | `error.code == "BAD_REQUEST"` | accepted |
| 4 | источник указывает на отсутствующий файл | 404 | `error.code == "SPEC_NOT_FOUND"` | accepted |
| 5 | спека не парсится как OpenAPI | 422 | `error.code == "SPEC_INVALID"` | accepted |
| 6 | спека — AsyncAPI | 501 | `error.code == "ASYNC_IMPACT_NOT_IMPLEMENTED"` | accepted |

```gherkin
@component @slice-03-diff-breaking-change
Feature: Assess the impact of a new provider spec version

  Scenario: breaking-изменение задевает потребителя
    Given запущенный сервис netlist и граф с ребром на операцию поставщика
    When клиент отправляет POST /impact с from_spec=v1 и to_spec=v2 (операция удалена)
    Then ответ 200 и affected_consumers содержит потребителя
    And details содержит субъект удалённой операции

  Scenario: non-breaking изменение никого не задевает
    When клиент отправляет POST /impact с from_spec=v1 и to_spec=v2 (добавлена операция)
    Then ответ 200 и affected_consumers пуст

  Scenario: запрос без источников спек
    When клиент отправляет POST /impact без from_spec/to_spec
    Then ответ 400 и error.code == "BAD_REQUEST"

  Scenario: файл спеки отсутствует
    When клиент отправляет POST /impact с spec_path на несуществующий файл
    Then ответ 404 и error.code == "SPEC_NOT_FOUND"

  Scenario: спека невалидна
    When клиент отправляет POST /impact с битой спекой
    Then ответ 422 и error.code == "SPEC_INVALID"

  Scenario: async-поставщик не поддержан
    When клиент отправляет POST /impact с AsyncAPI-спекой
    Then ответ 501 и error.code == "ASYNC_IMPACT_NOT_IMPLEMENTED"
```
