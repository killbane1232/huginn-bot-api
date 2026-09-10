# Huginn Bot API

Самостоятельный headless HTTP-сервис для ботов в сети Huginn. Сервис поднимает
один экземпляр Go-ядра, хранит его идентичность и историю в SQLite и открывает
авторизованный REST API для сообщений, файлов, событий, пиров и групп.
При каждом запуске ядро регистрирует Bot API как storage-пир `very_thick`; это
значение зафиксировано в коде и не переопределяется environment-переменными или
ранее сохранённой конфигурацией.

Ядро подключено как версионированная библиотечная зависимость без Git
submodule. При сборке последняя версия автоматически определяется через Go
module proxy, после чего скачивается готовая shared library из GitHub Release
`github.com/killbane1232/huginn-messenger`. Архив проверяется по опубликованному
`SHA256SUMS`, а адаптер `internal/core` загружает через `dlopen` совместимый
C ABI-вход с явным `peer_flag`.

## Быстрый запуск

```bash
git clone https://github.com/killbane1232/huginn-bot-api.git
cd huginn-bot-api
cp .env.example .env
# задайте BOT_API_TOKEN и HUGINN_USERNAME
docker compose up --build
```

При запуске контейнер исправляет владельца persistent volume `/app/data`, а
затем запускает Bot API от непривилегированного пользователя `bot` с UID
`10001`. Это позволяет повторно использовать volume, ранее созданный с
владельцем `root`, без удаления базы и ключей.

Локальная сборка требует Go 1.25+, GCC, Linux и `curl`:

```bash
make all

export BOT_API_TOKEN='replace-with-a-long-random-token'
export HUGINN_USERNAME='weather-bot'
export MUNINN_ADDR='https://your-muninn.example'
./build/huginn-bot-api
```

По умолчанию сервис слушает `:8081` и хранит состояние в `data/`.
Адрес Muninn обязателен в `MUNINN_ADDR` и имеет приоритет над ранее
сохранённым адресом в базе ядра; при смене ENV перезапустите бот. Для production
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

С ядром, поддерживающим устойчивую очередь, успешный ответ означает, что текст
и зашифрованные вложения уже сохранены в SQLite для фоновой доставки и повтора
после перезапуска. Ошибки получателя, чтения вложения и записи возвращаются в
HTTP-ответе. Это подтверждение приёма ботом, а не доставки получателю.

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

Принятые файлы сохраняются с правами `0640` в `BOT_API_UPLOAD_DIR`; путь также
остаётся в локальной истории. Новое ядро сохраняет зашифрованные чанки до
успешного ответа, поэтому повтор доставки не зависит от исходного файла.
При использовании старого release ядро может читать upload асинхронно;
очистку uploads должен выполнять оператор с учётом версии ядра.

Проверка адаптера с локально пересобранным ядром:

```bash
HUGINN_CORE_TEST_LIBRARY=/absolute/path/libhuginn_messenger.so go test ./...
```

Исправления доставки в исходниках ядра появятся в обычной сборке Bot API после
публикации соответствующего release либо при подключении локальной библиотеки.

## Конфигурация

| Переменная | Обязательность / default |
|---|---|
| `BOT_API_TOKEN` | обязательна |
| `HUGINN_USERNAME` | обязательна |
| `BOT_API_ADDR` | `:8081` |
| `HUGINN_CORE_LIBRARY` | `build/libhuginn_messenger.so` |
| `MUNINN_ADDR` | обязательна, URL с `http://` или `https://` |
| `HUGINN_DB_PATH` | `data/huginn.db` |
| `HUGINN_CHUNK_TTL` | `1w` |
| `BOT_API_UPLOAD_DIR` | `data/uploads` |
| `HUGINN_TURN_ADDR` | пусто |
| `HUGINN_TURN_USER` | пусто |
| `HUGINN_TURN_PASS` | пусто |

Если контейнер завершался с SQLite-ошибкой `unable to open database file (14)`,
пересоберите и перезапустите его: `docker compose up -d --build`. Новый
entrypoint восстановит права существующего `bot-data`; удалять volume не нужно.

## Разработка

```bash
go fmt ./...
go test ./...
make all
git diff --check
```

## Docker image

GitHub Actions workflow `.github/workflows/docker-publish.yml` проверяет Bot API,
собирает библиотеку Go-ядра, затем собирает образ для `linux/amd64` и
`linux/arm64`. На pull
request выполняется только сборка. Push в `main`, version-тег `v*.*.*` или
ручной запуск публикует образ в `ghcr.io/<owner>/<repository>` с тегами
`latest` для основной ветки, версией для Git-тега и неизменяемым `sha-*`.

Последняя опубликованная версия ядра подтягивается автоматически:

```bash
make all
```

Для воспроизводимой сборки можно явно задать версию; `GOPROXY` настраивает адрес
Go proxy:

```bash
HUGINN_CORE_VERSION=v0.2.0 make all
```
