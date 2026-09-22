# use-case — slice-03-diff-breaking-change

**UC-3: Оценить влияние новой версии спеки поставщика** · Level: user-goal · Trigger: CI поставщика отправляет `POST /impact {provider, to_version, from_spec, to_spec}`.

## Main success scenario (MSS)

1. CI поставщика изменил master-спеку и передаёт обе версии (старую и новую: `from_spec`/`to_spec`, путь|URL) — netlist спеки не хранит.
2. netlist загружает обе спеки, распарсивает (OpenAPI 3.x), вычисляет `from_version` из последнего вердикта поставщика в хранилище.
3. oasdiff-диф old→new даёт затронутые субъекты `METHOD /path` (удалённые операции, переименованные/удалённые поля, ужесточение типов).
4. По графу (`ConsumersOf(provider)`) netlist находит живущие рёбра с задетыми субъектами.
5. Ответ `200 BreakingChange{from_version, to_version, affected_consumers, details}`; заодно обновляется `LastCapturedHash(provider)` — свежесть provenance.

## Extensions (1 Extension = 1 error.code = 1 компонентный сценарий)

| # | Ветвь | error.code | HTTP |
|---|---|---|---|
| 3a | Запрос не соответствует схеме (нет provider/sources, source не exactly-one) | `BAD_REQUEST` | 400 |
| 3b | Источник спеки недоступен (файл отсутствует / HTTP не-2xx) | `SPEC_NOT_FOUND` | 404 |
| 3c | Спека не парсится как OpenAPI 3.x | `SPEC_INVALID` | 422 |
| 3d | Спека — AsyncAPI (async-поставщик) | `ASYNC_IMPACT_NOT_IMPLEMENTED` | 501 |

Happy-class — два различимых исхода: **breaking** (затронутые есть) и **non-breaking** (пустой список) — как compatible/incompatible у валидаторов.
