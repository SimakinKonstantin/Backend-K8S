# Учебный проект по бэкенду с Kubernetes

Предметная область - программы спортивных тренировок.

## Функционал
* Функционал управления программами тренировок;
* Функционал управления пользователями.

## Особенности
* Развертывание с помощью Kubernetes;
* Мониторинг через Grafana, Prometheus, дашборды настроены автоматически "из коробки" благодаря описанию в манифестах K8S;
* Аутентификация по JWT.

## Архитектура проекта
- **Gateway** - точка входа. Проксирует запросы к внутренним сервисам, выдаёт и проверяет JWT;
- **Сервис программ тренировок** - управляет программами тренировок. Предоставляет REST API, GraphQL API;
- **Сервис пользователей** - управляет пользователями приложения.

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
* `allPrograms` (query);
* `program` (query); 
* `programsCount` (query);
* `programFilter` (query);
* `updateProgram` (мутация);
* `addProgram` (мутация);
* `deleteProgram` (мутация);
* `deleteAllPrograms` (мутация).

Запросы можно отправлять через GraphQL Playground на порту 8010.

### JWT
JWT применяется в Gateway для аутентификации запросов. Роут `/login` проверяет существование пользователя через сервис пользователей и формирует JWT. Gateway проверяет токен перед проксированием на другие сервисы;

### MongoDB
Основное хранилище данных. Сервисы работают с разными коллекциями.

### Redis
Используется сервисом программ тренировок в качестве кэша программ, запрошенных по идентификатору. Используется стратегия кэширования Cache Aside.

### Apache Kafka
Обеспечивает асинхронное взаимодействие сервисов при подтверждении программы пользователем. Сервис программ тренировок публикует сообщение в топик `request`, передавая идентификатор программы и пользователя. Сервис пользователей читает этот топик, проверяет пользователя в MongoDB и публикует результат в топик `post`. Сервис программ тренировок читает ответ и сохраняет статус подтверждения программы. Таким образом асинхронное подтверждение программ не блокирует http запросы клиента.

### Kubernetes
Система развертывается в виде набора сущностей k8s. Манифесты указаны `./manifests`

## Инструкция по запуску
1. Запустить minikube: `minikube start`;
2. Применить манифесты: `kubectl apply -f .\manifests\ --recursive`;
3. Убедиться, что все поды готовы к использованию: `kubectl get pods`;
4. Хотя здесь настроен ingress, из-за особенностей minikube может понадобиться пробросить наружу порты: `kubectl port-forward service/gateway 8083:8083`;
5. Перейти на `http://localhost:8083/swagger/index.html`;
6. Чтобы просмотреть дашборды нужно пробросить порты до Grafana `kubectl port-forward service/grafana 3000:3000`;
7. Перейти на `http://localhost:3000`, в качестве кредов для входа использовать `admin` `admin`

## Скриншоты
<img width="672" height="345" alt="image" src="https://github.com/user-attachments/assets/71b47148-a601-416b-a806-1e3a0314163f" />
<p align="center"><strong>Поды готового приложения</strong></p><br>

<img width="902" height="733" alt="image" src="https://github.com/user-attachments/assets/636177f7-db35-4157-b78f-b53896fd5d8f" />
<p align="center"><strong>Swagger документация</strong></p><br>

<img width="1203" height="724" alt="image" src="https://github.com/user-attachments/assets/c8acfb73-fe30-4dd3-9631-f434694cbf21" />
<p align="center"><strong>Пример выполненного HTTP-запроса</strong></p><br>

<img width="1920" height="495" alt="image" src="https://github.com/user-attachments/assets/52377ee1-5589-4aee-a074-ec28c98b06b7" />
<p align="center"><strong>Пример выполненного GraphQL-запроса для получения только нужных клиенту полей</strong></p><br>

<img width="1008" height="396" alt="image" src="https://github.com/user-attachments/assets/852a552d-7f13-4e43-9749-de7649ef21cd" />
<p align="center"><strong>Автоматические настроенные через манифесты K8S дашборды</strong></p><br>

<img width="1919" height="879" alt="image" src="https://github.com/user-attachments/assets/c80731ab-de5b-497b-a194-a1e3d50e76cd" />
<p align="center"><strong>Дашборд метрик сервиса пользователей</strong></p><br>

<img width="1920" height="637" alt="image" src="https://github.com/user-attachments/assets/cd33806f-1bd0-43bf-817e-f3241d245592" />
<p align="center"><strong>Дашборд метрик Gateway сервиса</strong></p><br>

<img width="1920" height="873" alt="image" src="https://github.com/user-attachments/assets/c3f2df2b-d673-476a-b0c0-57299032daf5" />
<p align="center"><strong>Дашборд метрик сервиса программ тренировок</strong></p><br>
