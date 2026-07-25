# CLAUDE.md — pinout-netlist

Координатор графа связей сервисов. Концепт: [../pinout/README.md](../pinout/README.md). Эпик E2 в [бэклоге экосистемы](../pinout/backlog.md).

## Разработка: харнес izi (rationaldev-ai-sdlc-skills)

> Точка входа агента — [`AGENTS.md`](./AGENTS.md) (там же блок «Resume here»).

Разработку ведёт **харнес `izi`** ([rationaldev-ai-sdlc-skills](https://github.com/codemonstersteam/rationaldev-ai-sdlc-skills/)) — его скиллы, те же и в том же порядке, что и в `pinout-openapi`. Локальных копий не держим.

| Этап | Скилл | Ссылка |
|---|---|---|
| Документация | `documentation` | https://github.com/codemonstersteam/rationaldev-ai-sdlc-skills/tree/main/skills/lib/documentation |
| Проверка качества доков | `doc-quality-review` | https://github.com/codemonstersteam/rationaldev-ai-sdlc-skills/tree/main/skills/lib/doc-quality-review |
| Проектирование | `program-design` | https://github.com/codemonstersteam/rationaldev-ai-sdlc-skills/tree/main/skills/lib/program-design |
| Компонентные тесты | `component-tests` | https://github.com/codemonstersteam/rationaldev-ai-sdlc-skills/tree/main/skills/lib/component-tests |
| Реализация | `program-implementation` | https://github.com/codemonstersteam/rationaldev-ai-sdlc-skills/tree/main/skills/lib/program-implementation |

Ключевые правила — см. выжимку в `pinout-openapi/CLAUDE.md` (vertical slice, бизнес-логика ≠ I/O, конструкторы вместо guard, head-труба на `Result`, юниты только для логики).

## Особенность: это сетевой сервис

В отличие от `pinout-openapi` (CLI-инструмент), netlist — сетевой сервис с собственным API. Поэтому:
- собственный контракт лежит в `api-specification/openapi.yml` (контракт-первый);
- хранилище графа — за I/O-объектом `Store` (БД не торчит в зависимостях логики);
- внешние API (стаб) заменяются по реальному протоколу в компонентных тестах.

## Состояние сессии

Создан каркас + пакет проектирования верхнего уровня `docs/design/`. Реализация **не начата** (E2, после MVP `pinout-openapi`). Пер-срезовые контракты дозаполняются по `program-design`, когда стартует E2.
