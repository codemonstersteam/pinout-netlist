# pinout-netlist

Концепт экосистемы: [pinout](https://github.com/codemonstersteam/pinout) ([локально](../pinout/README.md)).

Координатор связей сервисов. Хранит граф `consumer → provider`, версии контрактов и принимает отчёты валидаторов ([`pinout-asyncapi`](../pinout-asyncapi), [`pinout-openapi`](../pinout-openapi)). Отвечает на вопрос, который одна пара спек не закрывает: **кто сломается, если поставщик так изменит контракт** — детект breaking-change во времени.

Название — отсылка к *netlist* в PCB-дизайне: формальное описание всех соединений схемы.

## Границы

- **Делает:** принимает отчёты валидаторов в [общем формате](../pinout-openapi/docs/report-format.md); строит граф пар consumer↔provider; хранит версии спек; диффит контракт поставщика против предыдущей версии и помечает затронутых живущих потребителей.
- **Не делает:** не валидирует совместимость пары (это валидаторы); не хранит сами спеки целиком (хранит ссылки/версии и вердикты).

## Зависимости

- Формат входа — канон [`pinout-openapi/docs/report-format.md`](../pinout-openapi/docs/report-format.md) (`schema_version`).
- Источники отчётов: эпики E0 (`pinout-asyncapi`) и E1 (`pinout-openapi`) в [бэклоге экосистемы](../pinout/backlog.md).

## Статус

📋 Запланирован (эпик E2). Здесь — каркас и пакет проектирования; реализация по TBD начинается после готовности общего формата отчёта (готов) и MVP `pinout-openapi`.

## Методология

Разработка по скиллам [service-template](https://github.com/ubik-life/service-template/) — см. [`CLAUDE.md`](./CLAUDE.md). Проектирование — [`docs/design/`](./docs/design/).

## Лицензия

Открытая лицензия (см. `LICENSE`).
