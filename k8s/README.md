# Инструкция по развёртыванию StatusHub в Kubernetes (Minikube)

Данный документ содержит исчерпывающий пошаговый листинг команд для установки окружения, сборки образов и развёртывания приложения **StatusHub — Server Inventory** и СУБД PostgreSQL в Minikube с использованием **ConfigMap** и **Secret**.

---

## Архитектура манифестов и сущностей

| Файл манифеста | Вид сущности (Kind) | Имя сущности | Описание и назначение |
| :--- | :--- | :--- | :--- |
| `k8s/configmap.yaml` | `ConfigMap` | `postgres-config`, `app-config` | Хранение открытых конфигураций (хост СУБД, порты, имя базы данных, имя пользователя, `PGDATA`). |
| `k8s/secret.yaml` | `Secret` | `postgres-secret`, `app-secret` | Защищённое хранение паролей (`POSTGRES_PASSWORD`, `DB_PASSWORD`) в формате Opaque/base64. |
| `k8s/postgres-pvc.yaml` | `PersistentVolumeClaim` | `postgres-pvc` | Выделение постоянного диска 1Gi (`ReadWriteOnce`) для персистентного хранения данных PostgreSQL. |
| `k8s/postgres-service.yaml` | `Service` | `postgres-service` | Внутренний `ClusterIP` сервис на порту `5432` для сетевого доступа из подов приложения по CoreDNS. |
| `k8s/postgres-deployment.yaml` | `Deployment` | `postgres-deployment` | Поддержание 1 реплики PostgreSQL 16 (с контейнером БД и sidecar exporter). Переменные внедряются из `postgres-config` и `postgres-secret`. |
| `k8s/app-service.yaml` | `Service` | `app-service` | Сервис типа `NodePort` (`nodePort: 30080`) для внешнего доступа к веб-интерфейсу и REST API с балансировкой трафика. |
| `k8s/app-deployment.yaml` | `Deployment` | `app-deployment` | Stateless-поды приложения StatusHub. Поддерживают масштабирование (`scale`). Параметры подключения к БД внедряются из `app-config` и `app-secret`. |

---

## Пошаговый листинг команд воспроизведения

### Шаг 1: Установка необходимых утилит

#### macOS (Homebrew):
```bash
# Установка Docker (если ещё не установлен)
brew install --cask docker

# Установка Minikube и kubectl
brew install minikube kubectl
```

#### Linux (Ubuntu/Debian):
```bash
# Установка Docker
sudo apt-get update && sudo apt-get install -y ca-certificates curl gnupg
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update && sudo apt-get install -y docker-ce docker-ce-cli containerd.io

# Установка Minikube
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube && rm minikube-linux-amd64

# Установка kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl && rm kubectl
```

---

### Шаг 2: Запуск Minikube и привязка к Docker-демону

```bash
# 1. Запуск одноузлового кластера Minikube
minikube start --driver=docker --cpus=2 --memory=4096

# 2. Переключение контекста Docker CLI на Docker-демон внутри виртуальной ноды Minikube
eval $(minikube docker-env)

# 3. Проверка статуса ноды
kubectl get nodes
```

---

### Шаг 3: Сборка и регистрация Docker-образа

```bash
# Сборка образа непосредственно в Docker-демоне Minikube
docker build -t status-hub:v3 -t status-hub:latest .

# (Если сборка производилась во внешнем Docker):
# minikube image load status-hub:v3
# minikube image load status-hub:latest

# Проверка наличия образа в кластере
minikube image ls | grep status-hub
```

---

### Шаг 4: Применение манифестов в кластере

Порядок применения учитывает цепочку зависимостей:
1. `ConfigMap` и `Secret` (конфигурации и учётные данные).
2. `PersistentVolumeClaim` (хранилище для БД).
3. `Service` и `Deployment` базы данных PostgreSQL.
4. `Service` и `Deployment` веб-приложения StatusHub.

```bash
# 1. Создание ConfigMap
kubectl apply -f k8s/configmap.yaml

# 2. Создание Secret
kubectl apply -f k8s/secret.yaml

# 3. Создание PVC для данных PostgreSQL
kubectl apply -f k8s/postgres-pvc.yaml

# 4. Развёртывание PostgreSQL (Service + Deployment)
kubectl apply -f k8s/postgres-service.yaml
kubectl apply -f k8s/postgres-deployment.yaml

# 5. Развёртывание StatusHub (Service + Deployment)
kubectl apply -f k8s/app-service.yaml
kubectl apply -f k8s/app-deployment.yaml
```

*Либо одновременное применение всех манифестов приложения:*
```bash
kubectl apply -f k8s/configmap.yaml \
              -f k8s/secret.yaml \
              -f k8s/postgres-pvc.yaml \
              -f k8s/postgres-service.yaml \
              -f k8s/postgres-deployment.yaml \
              -f k8s/app-service.yaml \
              -f k8s/app-deployment.yaml
```

---

### Шаг 5: Проверка статуса и подстановки Secret / ConfigMap

```bash
# Ожидание готовности развёртываний
kubectl rollout status deployment/postgres-deployment --timeout=90s
kubectl rollout status deployment/app-deployment --timeout=90s

# Просмотр всех ресурсов
kubectl get pods -o wide
kubectl get svc,pvc,configmap,secret

# Проверка переменных окружения в поде приложения (подстановка из CM и Secret)
kubectl describe pod -l app=status-hub | grep -A 15 "Environment:"

# Проверка чтения секрета
kubectl get secret app-secret -o jsonpath="{.data.DB_PASSWORD}" | base64 --decode && echo ""

# Просмотр логов приложения (успешный запуск и подключение к БД)
kubectl logs -l app=status-hub --tail=30
```

---

### Шаг 6: Доступ к веб-интерфейсу и тестирование функционала

```bash
# Вариант 1 (нативный Minikube NodePort):
minikube service app-service --url

# Вариант 2 (проброс порта на локальный хост):
kubectl port-forward svc/app-service 8080:8080

# Проверка списка серверов через REST API
curl -s http://localhost:8080/api/v1/servers | jq .

# Добавление нового сервера
curl -X POST http://localhost:8080/api/v1/servers \
  -H "Content-Type: application/json" \
  -d '{"name": "k8s-worker-01", "ip_address": "10.244.0.10", "environment": "production", "status": "active"}'
```

---

### Шаг 7: Масштабирование сервиса (Horizontal Scaling)

```bash
# Масштабирование до 3 реплик
kubectl scale deployment app-deployment --replicas=3

# Проверка статуса создания новых подов
kubectl rollout status deployment/app-deployment
kubectl get pods -l app=status-hub -o wide

# Проверка балансировки: все 3 пода работают со стабильным сервисом postgres-service
kubectl logs -l app=status-hub --tail=10
```

---

### Шаг 8: Очистка ресурсов

```bash
# Удаление всех созданных ресурсов приложения и БД
kubectl delete -f k8s/app-deployment.yaml \
               -f k8s/app-service.yaml \
               -f k8s/postgres-deployment.yaml \
               -f k8s/postgres-service.yaml \
               -f k8s/postgres-pvc.yaml \
               -f k8s/secret.yaml \
               -f k8s/configmap.yaml

# Остановка Minikube
minikube stop
```
