# StatusHub — Server Inventory System

> Сервис инвентаризации серверов и сетевой инфраструктуры. Разработан в соответствии с методологией **12-Factor App**.

## Стек технологий
* **Бэкенд**: Go 1.22+ (`net/http`, `log/slog`, `jackc/pgx/v5`)
* **База данных**: PostgreSQL 16
* **Фронтенд**: HTML5, Tailwind CSS, Vanilla JS
* **Контейнеризация**: Docker (Multi-stage), Docker Compose

---

## Быстрый запуск

### Запуск через Docker Compose (Рекомендуется)
```bash
docker compose up --build -d
```
После запуска откройте в браузере: **http://localhost:8080**

Остановка:
```bash
docker compose down
```

### Запуск тестов
```bash
go test -v ./...
```

---

## Архитектура проекта

```text
.
├── migrations/
│   └── 001_init.sql      # SQL-схема инициализации базы данных
├── static/
│   └── index.html        # Одностраничный интерфейс (дашборд)
├── Dockerfile            # Минималистичный multi-stage образ
├── docker-compose.yml    # Окружение БД + приложение
├── models.go             # Модели сущностей, валидация DTO
├── models_test.go        # Unit-тесты для валидаторов
├── db.go                 # Подключение к PostgreSQL и CRUD-запросы
├── main.go               # HTTP-сервер, REST API и Graceful Shutdown
├── Makefile              # Команды сборки и запуска
├── Отчёт.md              # Подробный отчёт по заданию и 12 факторам
└── README.md             # Документация проекта
```

---

## REST API Эндпоинты

| Метод | Путь | Описание |
| :--- | :--- | :--- |
| `GET` | `/api/v1/servers` | Список всех серверов (`?environment=`, `?search=`) |
| `GET` | `/api/v1/servers/{id}` | Получение сервера по ID |
| `POST` | `/api/v1/servers` | Создание нового сервера |
| `PUT` | `/api/v1/servers/{id}` | Обновление сервера |
| `DELETE` | `/api/v1/servers/{id}` | Удаление сервера |

---

## Конфигурация (12-Factor App)

Все настройки передаются через переменные окружения:
* `PORT` — сетевой порт (по умолчанию `8080`).
* `DATABASE_URL` — строка подключения к PostgreSQL.
* `LOG_FORMAT` — формат логов (`text` или `json`).
