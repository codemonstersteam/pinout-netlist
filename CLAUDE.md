# CLAUDE.md — pinout-netlist

Координатор графа связей сервисов. Концепт: [../pinout/README.md](../pinout/README.md). Эпик E2 в [бэклоге экосистемы](../pinout/backlog.md).

## Методология: скиллы service-template (по ссылке, не копируем)

> Точка входа агента — [`AGENTS.md`](./AGENTS.md) (там же блок «Resume here»).

Те же скиллы и порядок, что и в `pinout-openapi`. Источник — upstream [ubik-life/service-template](https://github.com/ubik-life/service-template/), локальных копий не держим. **Закреплённый коммит: `5ad8347c38ea7f07bd0620ebeef5030ca3431efd`** (2026-06-17) — при расхождении с `main` использовать его.

| Этап | Скилл | Ссылка |
|---|---|---|
| Документация | `documentation` | https://github.com/ubik-life/service-template/tree/main/skills/documentation |
| Проверка качества доков | `doc-quality-review` | https://github.com/ubik-life/service-template/tree/main/skills/doc-quality-review |
| Проектирование | `program-design` | https://github.com/ubik-life/service-template/tree/main/skills/program-design |
| Компонентные тесты | `component-tests` | https://github.com/ubik-life/service-template/tree/main/skills/component-tests |
| Реализация | `program-implementation` | https://github.com/ubik-life/service-template/tree/main/skills/program-implementation |

Ключевые правила — см. выжимку в `pinout-openapi/CLAUDE.md` (vertical slice, бизнес-логика ≠ I/O, конструкторы вместо guard, head-труба на `Result`, юниты только для логики).

## Особенность: это сетевой сервис

В отличие от `pinout-openapi` (CLI-инструмент), netlist — сетевой сервис с собственным API. Поэтому:
- собственный контракт лежит в `api-specification/openapi.yml` (контракт-первый);
- хранилище графа — за I/O-объектом `Store` (БД не торчит в зависимостях логики);
- внешние API (стаб) заменяются по реальному протоколу в компонентных тестах.

## Состояние сессии

Создан каркас + пакет проектирования верхнего уровня `docs/design/`. Реализация **не начата** (E2, после MVP `pinout-openapi`). Пер-срезовые контракты дозаполняются по `program-design`, когда стартует E2.
