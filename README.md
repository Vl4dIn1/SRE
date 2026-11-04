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

### Развёртывание в Kubernetes (Minikube)
Полный пошаговый листинг команд и описание манифестов вынесены в [k8s/README.md](file:///Users/vladislav/Documents/SRE/k8s/README.md).

Быстрый запуск:
```bash
# 1. Запуск Minikube и переключение на кластерный Docker
minikube start --driver=docker
eval $(minikube docker-env)

# 2. Сборка образа
docker build -t status-hub:v3 -t status-hub:latest .

# 3. Применение конфигураций, секретов, хранилища и сервисов
kubectl apply -f k8s/configmap.yaml \
              -f k8s/secret.yaml \
              -f k8s/postgres-pvc.yaml \
              -f k8s/postgres-service.yaml \
              -f k8s/postgres-deployment.yaml \
              -f k8s/app-service.yaml \
              -f k8s/app-deployment.yaml

# 4. Открытие сервиса в браузере
minikube service app-service --url
```

---

## Архитектура проекта

```text
.
├── k8s/                  # Декларативные манифесты Kubernetes
│   ├── configmap.yaml    # ConfigMap: нечувствительные параметры БД и приложения
│   ├── secret.yaml       # Secret: защищённые пароли БД и приложения
│   ├── postgres-pvc.yaml # PersistentVolumeClaim: постоянное хранилище PostgreSQL
│   ├── postgres-service.yaml    # Service (ClusterIP): внутренняя сеть БД
│   ├── postgres-deployment.yaml # Deployment: база данных PostgreSQL 16
│   ├── app-service.yaml         # Service (NodePort): внешний доступ на 30080
│   ├── app-deployment.yaml      # Deployment: приложение StatusHub
│   └── README.md         # Подробный листинг команд воспроизведения
├── migrations/
│   └── 001_init.sql      # SQL-схема инициализации базы данных
├── static/
│   └── index.html        # Одностраничный интерфейс (дашборд)
├── Dockerfile            # Минималистичный multi-stage образ
├── docker-compose.yml    # Окружение БД + приложение
├── models.go             # Модели сущностей, валидация DTO
├── models_test.go        # Unit-тесты для валидаторов и конфигураций
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
