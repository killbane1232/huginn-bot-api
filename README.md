# Huginn Bot API

Самостоятельный headless HTTP-сервис для ботов в сети Huginn. Сервис поднимает
один экземпляр Go-ядра, хранит его идентичность и историю в SQLite и открывает
авторизованный REST API для сообщений, файлов, событий, пиров и групп.
При каждом запуске ядро регистрирует Bot API как storage-пир `very_thick`; это
значение зафиксировано в коде и не переопределяется environment-переменными или
ранее сохранённой конфигурацией.

Ядро подключено в `third_party/huginn-messenger` как Git submodule. Bot API не
импортирует его `internal`-пакеты: при сборке ядро превращается в shared library,
а адаптер `internal/core` загружает через `dlopen` совместимый C ABI-вход с
явным `peer_flag`.

## Быстрый запуск

```bash
git clone --recurse-submodules https://github.com/killbane1232/huginn-bot-api.git
cd huginn-bot-api
cp .env.example .env
# задайте BOT_API_TOKEN и HUGINN_USERNAME
docker compose up --build
```

Локальная сборка требует Go 1.25+, GCC и Linux:

```bash
git submodule update --init --recursive
make all

export BOT_API_TOKEN='replace-with-a-long-random-token'
export HUGINN_USERNAME='weather-bot'
./build/huginn-bot-api
```

По умолчанию сервис слушает `:8081`, подключается к
`https://muninn.evil-bread.ru` и хранит состояние в `data/`. Для production
следует использовать HTTPS reverse proxy и не передавать токен в query string.

## Авторизация

`GET /healthz` открыт для health checks. Все `/api/v1/*` требуют заголовок:

```text
Authorization: Bearer <BOT_API_TOKEN>
```

## Методы

| Метод | Маршрут | Назначение |
|---|---|---|
| `GET` | `/api/v1/me` | Идентичность бота |
| `GET` | `/api/v1/peers?q=` | Пиры или поиск пиров |
| `GET` | `/api/v1/messages?chat_id=&limit=&offset=` | История чата |
| `POST` | `/api/v1/messages` | Отправить текст |
| `POST` | `/api/v1/files` | Отправить файл (`multipart/form-data`) |
| `GET` | `/api/v1/updates?timeout=25` | Long polling события до 30 секунд |
| `GET` | `/api/v1/groups` | Группы |
| `POST` | `/api/v1/groups` | Создать группу |
| `POST` | `/api/v1/groups/invitations` | Пригласить участника |
| `POST` | `/api/v1/messages/read` | Отметить сообщение прочитанным |

Пример отправки сообщения:

```bash
curl -H "Authorization: Bearer $BOT_API_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"chat_id":"peer-id-or-login:signature","text":"hello"}' \
  http://localhost:8081/api/v1/messages
```

Пример получения следующего события:

```bash
curl -H "Authorization: Bearer $BOT_API_TOKEN" \
  'http://localhost:8081/api/v1/updates?timeout=25'
```

События имеют типы ядра `message`, `peers` и `file_ready`. При таймауте сервер
возвращает `{"event":null}`.

Файл отправляется так:

```bash
curl -H "Authorization: Bearer $BOT_API_TOKEN" \
  -F chat_id='peer-id-or-login:signature' \
  -F caption='report' \
  -F file=@report.pdf \
  http://localhost:8081/api/v1/files
```

Принятые файлы сохраняются с правами `0640` в `BOT_API_UPLOAD_DIR`, потому что
ядро может читать их асинхронно после HTTP-ответа. Очистку старых uploads должен
выполнять оператор после подтверждения доставки.

## Конфигурация

| Переменная | Обязательность / default |
|---|---|
| `BOT_API_TOKEN` | обязательна |
| `HUGINN_USERNAME` | обязательна |
| `BOT_API_ADDR` | `:8081` |
| `HUGINN_CORE_LIBRARY` | `build/libhuginn_messenger.so` |
| `MUNINN_ADDR` | `https://muninn.evil-bread.ru` |
| `HUGINN_DB_PATH` | `data/huginn.db` |
| `HUGINN_CHUNK_TTL` | `1w` |
| `BOT_API_UPLOAD_DIR` | `data/uploads` |
| `HUGINN_TURN_ADDR` | пусто |
| `HUGINN_TURN_USER` | пусто |
| `HUGINN_TURN_PASS` | пусто |

## Разработка

```bash
go fmt ./...
go test ./...
make all
git diff --check
```

Обновление ядра выполняется отдельным изменением gitlink:

```bash
git -C third_party/huginn-messenger fetch origin
git -C third_party/huginn-messenger checkout <tested-commit>
git add third_party/huginn-messenger
```
