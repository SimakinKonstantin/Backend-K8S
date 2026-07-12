# Учебный проект по бэкенду

Предметная область - программы спортивных тренировок.

## Архитектура проекта
- **Gateway** - точка входа. Проксирует запросы к внутренним сервисам, выдаёт и проверяет JWT;
- **Сервис программ тренировок** - управляет программами тренировок. Предоставляет REST API, GraphQL API;
- **Сервис пользователей** — управляет пользователями приложения.

## Технологии

- **Язык:** Go;
- **API:** HTTP, GraphQL, Swagger/OpenAPI (через `swaggo`);
- **Аутентификация:** JWT;
- **Постоянное хранилище данных:** MongoDB;
- **Кэш:** Redis;
- **Брокер сообщений:** Apache Kafka;
- **Мониторинг:** Prometheus, Grafana;
- **Контейнеризация и оркестрация:** Kubernetes, Minikube.

### GraphQL
GraphQL предоставляет API сервиса программ тренировок в дополнение к HTTP. Реализует запросы:
* `allPrograms`;
* `program`; 
* `programsCount`;
* `programFilter`.
Также реализованы мутации:
* `updateProgram`;
* `addProgram`;
* `deleteProgram`;
* `deleteAllPrograms`.

Запросы можно отправлять через GraphQL Playground на порту 8010.

### JWT
JWT применяется в Gateway для аутентификации запросов. Роут `/login` проверяет существование пользователя через сервис пользователей и формирует JWT. Gateway проверяет токен перед проксированием на другие сервисы;

### MongoDB
Основное хранилище данных. Сервисы работают с разными коллекциями.

### Redis
Используется сервисом программ тренировок в качестве кэша программ, запрошенных по идентификатору. Используется стратегия кэширования Cache Aside.

### Apache Kafka
Обеспечивает асинхронное взаимодействие сервисов при подтверждении программы пользователем. Сервис программ тренировок публикует сообщение в топик `request`, передавая идентификатор программы и пользователя. Сервис пользователей читает этот топик, проверяет пользователя в MongoDB и публикует результат в топик `post`. Сервис программ тренировок читает ответ и сохраняет статус подтверждения программы. Таким образом асинхронное подтверждение программ не блокирует http запросы клиента.

## Взаимодействие сервисов

При создании программы сервис программ отправляет в Kafka запрос на подтверждение пользователя. Сервис пользователей читает запрос, проверяет наличие пользователя и публикует результат. Сервис программ получает ответ и обновляет статус программы. Такой обмен выполняется асинхронно и не блокирует HTTP-запрос клиента.

Gateway перед перенаправлением защищённых запросов проверяет JWT и существование пользователя через Users service.

## API

### Gateway

Базовый адрес: `http://<gateway-address>:8083`.

| Метод | Эндпоинт | Назначение |
| --- | --- | --- |
| `GET` | `/login?login={login}` | Выдаёт JWT для существующего пользователя. |
| `GET` | `/gateway/programs` | Получить все программы. |
| `GET` | `/gateway/programs/?id={id}` | Получить программу по идентификатору. |
| `POST` | `/gateway/programs/` | Создать программу. |
| `PATCH` | `/gateway/programs/?id={id}` | Изменить программу. |
| `DELETE` | `/gateway/programs/?id={id}` | Удалить программу. |
| `DELETE` | `/gateway/programs` | Удалить все программы. |
| `GET` | `/gateway/programs/count` | Получить число программ. |
| `GET` | `/gateway/users` | Получить всех пользователей. |
| `GET` | `/gateway/users/?id={id}` | Получить пользователя по идентификатору. |
| `POST` | `/gateway/users/` | Создать пользователя. |
| `PATCH` | `/gateway/users/?id={id}` | Изменить пользователя. |
| `DELETE` | `/gateway/users/?id={id}` | Удалить пользователя. |
| `DELETE` | `/gateway/users` | Удалить всех пользователей. |
| `GET` | `/gateway/users/count` | Получить число пользователей. |
| `GET` | `/health` | Проверить доступность зависимых сервисов. |
| `GET` | `/metrics` | Метрики Prometheus. |
| `GET` | `/swagger/` | Swagger UI. |

Для защищённых маршрутов передайте полученный JWT в заголовке `Authorization`. Текущая реализация ожидает непосредственно значение токена, без префикса `Bearer`.

### Programs service

REST API сервиса доступно на порту `8081`:

- `GET /programs`, `GET /programs/?id={id}`;
- `POST /programs/`, `PATCH /programs/?id={id}`;
- `DELETE /programs`, `DELETE /programs/?id={id}`;
- `GET /programs/count`;
- `GET /health`, `GET /metrics`, `GET /swagger/`.

GraphQL API запущено на порту `8010`:

- `GET /` — GraphQL Playground;
- `GET` или `POST /query` — GraphQL endpoint.

В GraphQL-схеме доступны запросы `allPrograms`, `program`, `programsCount`, `programFilter` и мутации `addProgram`, `updateProgram`, `deleteProgram`, `deleteAllPrograms`.

### Users service

REST API сервиса доступно на порту `8080`:

- `GET /users`, `GET /users/?id={id}`;
- `POST /users/`, `PATCH /users/?id={id}`;
- `DELETE /users`, `DELETE /users/?id={id}`;
- `GET /users/count`;
- `GET /users/validness?login={login}` — проверка существования пользователя;
- `GET /health`, `GET /metrics`, `GET /swagger/`.

## Развёртывание в Minikube

Требуются установленные `minikube` и `kubectl`.

```bash
minikube start
minikube addons enable ingress
kubectl apply -f ./manifests --recursive
kubectl get pods
```

Дождитесь, пока все поды перейдут в состояние `Running`. Для диагностики используйте:

```bash
kubectl get services
kubectl get ingress
kubectl get pods
```

Сервисы имеют тип `NodePort`, поэтому получить их адреса можно через Minikube:

```bash
minikube service gateway --url
minikube service microservice --url
minikube service user-service --url
minikube service prometheus --url
minikube service grafana --url
```

Для доступа к маршрутам Ingress включите туннель в отдельном терминале:

```bash
minikube tunnel
```

Затем добавьте IP-адрес из `minikube ip` в файл `hosts` вместе с именами `gateway.local`, `programs.local` и `users.local`, которые определены в `manifests/ingress.yaml`.

Grafana открывается по адресу, выведенному командой `minikube service grafana --url`. Учётные данные: `admin` / `admin`.

## Локальный запуск через Docker Compose

Для запуска всего окружения в Docker выполните:

```bash
docker compose up --build
```

После запуска доступны Gateway на порту `8083`, Programs service на `8081`, GraphQL на `8010`, Users service на `8080`, Prometheus на `9090` и Grafana на `3000`.



## Структура каталогов

- `Gateway/` — исходный код API Gateway;
- `Microservices/` — исходный код сервиса программ и GraphQL-схема;
- `UserMicroservice/` — исходный код сервиса пользователей;
- `manifests/` — Kubernetes-манифесты сервисов, хранилищ, Kafka, мониторинга и Ingress;
- `prometheus/` — конфигурация Prometheus для Docker Compose;
- `compose.yaml` — локальная конфигурация Docker Compose.
