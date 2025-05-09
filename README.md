# Оглавление
- [Описание проекта](#описание-проекта)
- [Архитектура приложения](#архитектура-приложения)
  - [Оркестратор](#оркестратор)
  - [Агент](#агент)
- [Схема работы](#схема-работы)
- [Запуск проекта](#запуск-проекта)
  - [Переменные окружения](#переменные-окружения)
  - [Запуск оркестратора](#запуск-оркестратора)
  - [Запуск агента](#запуск-агента)
  - [Запуск через Docker Compose](#запуск-через-docker-compose)
  - [Запуск тестов](#запуск-тестов)
- [API Endpoints](#api-endpoints)
  - [POST /api/v1/calculate](#post-apiv1calculate)
  - [GET /api/v1/expressions](#get-apiv1expressions)
  - [GET /api/v1/expressions/:id](#get-apiv1expressionsid)
  - [GET /internal/task](#get-internaltask)
  - [POST /internal/task](#post-internaltask)
# Распределённый вычислитель арифметических выражений
## Описание проекта
Приложение позволяет пользователю вводить арифметические выражения, которые вычисляются в фоновом режиме с использованием нескольких вычислительных агентов. Каждая операция (сложение, вычитание, умножение, деление) выполняется отдельно, что позволяет масштабировать систему путём добавления новых вычислительных мощностей.

## Архитектура приложения
### Оркестратор:
- Сервер, который принимает арифметические выражения, разбивает их на задачи и управляет их выполнением.
- Предоставляет API для добавления выражений, получения статуса и результатов вычислений.
### Агент:
- Демон, который получает задачи от оркестратора, выполняет вычисления и возвращает результаты.
- Может запускать несколько горутин для параллельного выполнения задач.

### Схема работы
Оркестратор и агенты запускаются в отдельных терминалах. Агент с периодичностью 3 секунды опрашивает оркестратор (GET **/internal/task), 
не появилась ли какая-либо задача для вычисления. Если появилась, делает вычисление и записывает результат (POST **/internal/task).
Чтобы увидеть работу агентов, нужно будет создать выражение с помощью POST **/api/v1/calculate.

## Запуск проекта
Скопируйте проект
```commandline
git clone https://github.com/pliliya111/go_final_sprint.git
```
Установите необходимые зависимости
```commandline
go mod tidy
```
### Переменные окружения

TIME_ADDITION_MS — время выполнения операции сложения (в миллисекундах).

TIME_SUBTRACTION_MS — время выполнения операции вычитания (в миллисекундах).

TIME_MULTIPLICATIONS_MS — время выполнения операции умножения (в миллисекундах).

TIME_DIVISIONS_MS — время выполнения операции деления (в миллисекундах).

COMPUTING_POWER — количество горутин для выполнения задач (по умолчанию 1).

#### Пример настройки переменных окружения
```commandline
export TIME_ADDITION_MS=1000
export TIME_SUBTRACTION_MS=1000
export TIME_MULTIPLICATIONS_MS=2000
export TIME_DIVISIONS_MS=2000
export COMPUTING_POWER=4
```

### Запуск оркестратора
```
go run cmd/orchestrator/main.go 
```
### Запуск агента
```
go run cmd/agent/main.go 
```
### Запуск тестов
```
go test -v ./...
go test -v ./internal/handler
go test -v ./internal/calculator
```

## Запуск через Docker Compose
```
docker-compose up
```

## API Endpoint

### POST /api/v1/calculate
#### Примеры запросов
1) Регистрация пользователя
```
curl --location 'localhost:8080/api/v1/register' \
--header 'Content-Type: application/json' \
--data '{
  "name": "user_1",
  "password": "123"
}'
```
2) Логин пользователя
```
curl --location 'localhost:8080/api/v1/login' \
--header 'Content-Type: application/json' \
--data '{
  "name": "user_1",
  "password": "123"
}'
```
3) Добавление выражения для вычисления
```
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer <Token>' \
--data '{
  "expression": "2+2*2"
}'
```
Ответ:
- Код ответа 201
- Тело ответа:
```commandline
{"id":"a91f6bf8-2008-4b00-b44b-8ac81534e135"}
```
4) Получение списка выражений
```commandline
curl --location 'localhost:8080/api/v1/expressions' \
--header 'Authorization: Bearer <Token>' 
```
Ответ:
- Код ответа: 200
- Тело ответа:
```
{
  "expressions": [
    {
      "id": "db035ace-6fa0-4f7a-97fa-f37f08cb3761",
      "status": "in_progress",
      "result": null
    },
    {
      "id": "ad6c3f6c-787e-4d94-843b-63f60a013f86",
      "status": "completed",
      "result": 6
    }
  ]
}
```
5) Получение выражения по идентификатору
```
curl --location 'localhost:8080/api/v1/expressions/db035ace-6fa0-4f7a-97fa-f37f08cb3761'
```
Ответ:
- Код ответа: 200
- Тело ответа:
```
{
  "expression": {
    "id": "db035ace-6fa0-4f7a-97fa-f37f08cb3761",
    "status": "in_progress",
    "result": null
  }
}
```
Ответ:
Код ответа: 404
Тело ответа:
```
{"error": "task not found"}
```
6) Получение задачи для выполнения (для агентов)
```
curl --location 'localhost:8080/internal/task'
```
Ответ:
Код ответа: 200
Тело ответа:
```
{
  "task": {
    "id": "ad6c3f6c-787e-4d94-843b-63f60a013f86",
    "arg1": "2",
    "arg2": "2",
    "operation": "*",
    "operation_time": 2000
  }
}
```
7) Отправка результата выполнения задачи (для агентов)
```
curl --location 'localhost:8080/internal/task' \
--header 'Content-Type: application/json' \
--data '{
  "id": "cd7f328a-31a6-4d57-8b0c-796cbfe42316",
  "result": 6
}'
```
Ответ:
Код ответа: 200
Тело ответа:
```
{"message": "result submitted"}
```
Ответ:
Код ответа: 404
Тело ответа:
```
{"error": "task not found"}
```
