# Тикет — привести дизайн-пакет к канону отчёта `1.1` (до старта E2)

**Приоритет:** 🔴 №1 в репозитории, **до запуска харнес-прогона E2**.
**Родительский разбор:** `pinout/debt/report-canon-fork.md` (решение оператора — вариант A, 2026-07-25).
**Зависит от:** `pinout-openapi/debt/01-report-canon-doc.md` — канон
`pinout-openapi/docs/report-format.md` должен быть написан и смержен. Он источник истины формата.

## Состояние репозитория — разработки ещё нет

Это не правка работающего сервиса. На момент тикета:

- **Go-кода нет**: ни `go.mod`, ни одного `.go`-файла. Последние коммиты — каркас и документация
  (`eea7129`, `509ee70`, `4f6f04b`);
- `api-specification/openapi.yml` — **черновик**, `info.version: 0.1.0`, **`x-frozen` отсутствует**:
  контракт не заморожен, править его можно свободно;
- дизайн-пакет частичный: `docs/design/{intent,slices,messages,backlog}.md`; в `slices.md` прямо
  записано, что контракты срезов дозаполняются по `program-design` **при старте E2**;
- E2 в бэклоге экосистемы — статус 📋, реализация после E1/E0.

Отсюда объём тикета: **только документы и черновая спека**. Ни реализации, ни тестов, ни миграций
нет и не создаётся. Это самый дешёвый момент для правки — после старта E2 те же изменения задевали
бы замороженный контракт, код хранилища и фикстуры.

## Дефект

Отчёты валидаторов — единственный вход этого сервиса (срез 01 `ingest-report`). Дизайн-пакет
описывает вход **по форме, которую ни один валидатор не печатает**:

- `api-specification/openapi.yml:65` — схема `ValidatorReport` с
  `required: [schema_version, validator, interaction, consumer, provider, compatible, verdicts]`.
  Валидаторы печатают плоский `errors[]` без `verdicts[]` и без `provider{}`;
- `docs/design/messages.md:35` — `VerdictRecord.At` берётся из `report.generated_at`, поля в
  отчётах нет;
- `Edge{Consumer, Provider, Subject, Interaction}` нечем заполнить: имени потребителя и типа
  взаимодействия в отчёте нет, субъект спрятан в прозе `location`.

Плюс 8 битых ссылок на удалённый канон `pinout-openapi/docs/report-format.md`.

Канон `1.1` закрывает нехватку полей со стороны валидаторов; этот тикет приводит к нему **сторону
потребителя**, пока она существует только на бумаге.

## Работы

### 1. Черновая спека

- [ ] `api-specification/openapi.yml` — переписать `ValidatorReport` под канон `1.1`:
      `required: [schema_version, validator, interaction, consumer, compatible, provenance, errors]`;
      `errors[]{code, message, subject, location, details, context}`; `generated_at`;
      `uncovered_operations` / `uncovered_channels` — опциональные, по типу валидатора;
- [ ] **удалить схему `provider{}`/`Party` в части поставщика** — идентичность поставщика несёт
      `provenance{provider, provider_version, captured_hash}`, дублировать её нельзя;
- [ ] один reader на оба валидатора: различие sync/async только в `validator` / `interaction` /
      имени поля `uncovered_*` и в словаре кодов.

### 2. Модель данных

- [ ] `docs/design/messages.md` — перепривязать поля к канону `1.1`:

| поле модели | источник в отчёте `1.1` |
|---|---|
| `Edge.Consumer` | `consumer.name` |
| `Edge.Provider` | `provenance.provider` |
| `Edge.Subject` | `errors[].subject` (и перечень субъектов из конфига потребителя) |
| `Edge.Interaction` | `interaction` |
| `VerdictRecord.At` | `generated_at` |
| `VerdictRecord.ProviderVersion` | `provenance.provider_version` |
| `VerdictRecord.Errors` | `errors[]` |
| свежесть provenance | `provenance.captured_hash` |

- [ ] **зафиксировать следствие варианта A:** отчёт плоский, поэтому один отчёт раскладывается в N
      рёбер по `errors[].subject`, а `compatible` в отчёте — **агрегатный**. Совместимость
      конкретного ребра выводится как «нет ошибки с этим субъектом», а не читается из отчёта.
      Записать это явно в `messages.md` — иначе при реализации E2 родится баг «все рёбра красные,
      если сломано одно»;
- [ ] отметить `provenance.captured_hash` как ключ авто-перепроверки forward по свежести
      (пункт бэклога E2) — раньше такого источника не было.

### 3. Битые ссылки (8)

Заменить ссылки на восстановленный канон `pinout-openapi/docs/report-format.md`:

- [ ] `README.md:11`, `README.md:16`
- [ ] `AGENTS.md:7`
- [ ] `docs/design/intent.md:5`
- [ ] `api-specification/openapi.yml:65` (комментарий над схемой)
- [ ] `docs/design/backlog.md:25`
- [ ] `docs/design/messages.md:29`, `docs/design/messages.md:35`

## Границы

Не проектировать срезы, не заводить `go.mod`, не писать код и тесты, не замораживать спеку — всё
это делает прогон E2 по `program-design`. Здесь только приведение существующих документов и
черновой спеки к канону.

## Приёмка

- `ValidatorReport` в `openapi.yml` **валидирует реальный вывод обоих валидаторов** версии `1.1`
  (проверить на фактическом выходе `pinout-openapi validate` после закрытия его тикета 2/2);
- `provider{}` как отдельная сущность в схеме отчёта отсутствует;
- `messages.md` содержит таблицу привязки полей и явную запись про агрегатный `compatible`;
- битых ссылок на `report-format.md` в репозитории — **ноль** (`grep -rn "report-format"`);
- `go.mod` и `.go`-файлы не появились — объём тикета не расширен.
