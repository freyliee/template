# Лабораторная работа 2

## Запуск

```bash
tripgoctl cluster start
tripgoctl environment start
tripgoctl connect
make migrate
make run
```

## Тесты и проверка кода

```bash
make build
make generate
make lint
make test
make test-integration
```

## Архитектура

```
cmd/trip-service: config, container
container: HTTP-обработчики, usecase, repository, infra
HTTP: интерфейсы usecase, domain/model, generated
usecase: domain/model, интерфейсы repository/transaction/clock/observer
repository: domain, pgx/squirrel, infra/postgres/transaction
infra: PostgreSQL, время, slog и OpenTelemetry
```

`main` читает конфигурацию, вызывает `container.New`, запускает HTTP-сервер
и управляет остановкой. Конкретные зависимости собираются в `container`.
Пул создаётся фабрикой `infra/postgres.NewPool`, вызываемой из контейнера.

Каждому действию соответствуют пакеты `usecase/trip/create`, `get`, `finish`
и одноимённые HTTP-пакеты. Неэкспортируемые интерфейсы находятся в `contract.go`
у потребителя. Например, для получения поездки нужен только `GetTrip`,
а не весь репозиторий.

`domain/trip` содержит поездку, входные данные, статусы, валидацию и `Finish`.
Домен не импортирует HTTP, pgx, OpenAPI или OTel. `model.CreateTripInput` - имя
входного DTO для usecase. Его структура совпадает с `trip.Data`, поэтому 
использован алиас без второго одинакового набора полей. HTTP-конвертеры
явно преобразуют generated-модели в доменные и обратно.