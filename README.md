# GophKeeper

README по сборке и запуску серверной и клиентской частей приложения.

## Что входит в проект

В текущем виде проект состоит из двух приложений:

- **gRPC-сервер** — подключается к PostgreSQL, запускает миграции, и поднимает `AuthService` и `StorageService`.
- **HTTP-клиент / HTTP API gateway** — поднимает HTTP-сервер и ходит в gRPC-сервер по адресу из конфига.

Оба приложения при старте печатают `buildVersion`, `buildDate`, `buildCommit`.

---

## Требования

- Go 1.22+
- PostgreSQL
- настроенные переменные окружения, флаги или JSON-конфиг

---

## Сборка

Ниже команды приведены в двух вариантах:

- через типичную структуру `./cmd/server` и `./cmd/client`
- через текущую директорию `.` — если `main.go` лежит прямо в корне нужного main-пакета

### Сборка серверной части

```bash
go build -ldflags "-X main.buildVersion=1.0.0 -X 'main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)' -X main.buildCommit=$(git rev-parse --short HEAD)" -o gophkeeper-server ./cmd/server
```

Если main-пакет сервера лежит в текущей папке:

```bash
go build -ldflags "-X main.buildVersion=1.0.0 -X 'main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)' -X main.buildCommit=$(git rev-parse --short HEAD)" -o gophkeeper-server .
```

### Сборка клиентской HTTP-части

```bash
go build -ldflags "-X main.buildVersion=1.0.0 -X 'main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)' -X main.buildCommit=$(git rev-parse --short HEAD)" -o gophkeeper-client ./cmd/client
```

Если main-пакет клиента лежит в текущей папке:

```bash
go build -ldflags "-X main.buildVersion=1.0.0 -X 'main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)' -X main.buildCommit=$(git rev-parse --short HEAD)" -o gophkeeper-client .
```

### Для Windows PowerShell

#### Сервер

```powershell
$env:BUILD_DATE=(Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$env:BUILD_COMMIT=(git rev-parse --short HEAD)
go build -ldflags "-X main.buildVersion=1.0.0 -X main.buildDate=$env:BUILD_DATE -X main.buildCommit=$env:BUILD_COMMIT" -o gophkeeper-server.exe .
```

#### Клиент

```powershell
$env:BUILD_DATE=(Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$env:BUILD_COMMIT=(git rev-parse --short HEAD)
go build -ldflags "-X main.buildVersion=1.0.0 -X main.buildDate=$env:BUILD_DATE -X main.buildCommit=$env:BUILD_COMMIT" -o gophkeeper-client.exe .
```

---

## Конфигурация

И сервер, и клиент читают конфиг с одинаковым приоритетом:

```text
ФЛАГИ > ENV > CONFIG(JSON) > DEFAULTS
```

---

## Серверная часть

### Что делает сервер при старте

Сервер:

1. печатает build-информацию;
2. загружает конфигурацию;
3. подключается к PostgreSQL;
4. выполняет миграции;
5. создает `DBStore`, JWT manager и ключ шифрования;
6. поднимает gRPC-сервер;
7. регистрирует `AuthService` и `StorageService`.

### Флаги сервера

```bash
-g   # адрес gRPC-сервера
-d   # DSN для подключения к базе данных
-k   # ключ для шифрования данных
-a   # ключ для JWT / auth
-df  # таймаут удаления зависших файлов
-t   # TTL токена пользователя
-c   # путь к JSON-конфигу
```

### Переменные окружения сервера

```bash
DSN
GRPC_SERVER_ADDRESS
CRYPTO_KEY
AUTH_KEY
DELETE_TIMEOUT
TOKEN_TTL
CONFIG
```

### Важные замечания по ключам сервера

- `CRYPTO_KEY` должен быть в Base64.
- После декодирования ключ должен быть ровно **32 байта**.
- `AUTH_KEY` используется для JWT.

### Пример запуска сервера через ENV

#### Linux / macOS

```bash
export DSN='postgres://postgres:postgres@localhost:5432/gophkeeper?sslmode=disable'
export GRPC_SERVER_ADDRESS=':5051'
export CRYPTO_KEY='BASE64_КЛЮЧ_ИЗ_32_БАЙТ'
export AUTH_KEY='super-secret-auth-key'
export DELETE_TIMEOUT='24'
export TOKEN_TTL='24'

./gophkeeper-server
```

#### Windows PowerShell

```powershell
$env:DSN='postgres://postgres:postgres@localhost:5432/gophkeeper?sslmode=disable'
$env:GRPC_SERVER_ADDRESS=':5051'
$env:CRYPTO_KEY='BASE64_КЛЮЧ_ИЗ_32_БАЙТ'
$env:AUTH_KEY='super-secret-auth-key'
$env:DELETE_TIMEOUT='24'
$env:TOKEN_TTL='24'

./gophkeeper-server.exe
```

### Пример JSON-конфига сервера

`config.server.json`

```json
{
  "dsn": "postgres://postgres:postgres@localhost:5432/gophkeeper?sslmode=disable",
  "grpc_server_address": ":5051",
  "crypto_key": "BASE64_КЛЮЧ_ИЗ_32_БАЙТ",
  "delete_file_timeout": 24,
  "auth_key": "super-secret-auth-key",
  "token_TTL": 24
}
```

Запуск:

```bash
./gophkeeper-server -c ./config.server.json
```

или:

```bash
CONFIG=./config.server.json ./gophkeeper-server
```

### Генерация Base64-ключа

```bash
openssl rand -base64 32
```

Важно: нужен ключ, который после Base64-декодирования дает **ровно 32 байта**.

---

## Клиентская HTTP-часть

### Что делает клиент при старте

Клиентское приложение:

1. печатает build-информацию;
2. загружает конфигурацию клиента;
3. поднимает HTTP-сервер на `Host`;
4. использует `GrpcServer` как адрес backend gRPC-сервера.

По сути это HTTP-обвязка над gRPC-сервисами.

### Флаги клиента

```bash
-a   # адрес запуска HTTP-сервера
-g   # адрес gRPC-сервера
-c   # путь к JSON-конфигу
```

### Переменные окружения клиента

```bash
HOST
GRPC_SERVER_ADDRESS
CONFIG
```

### Значения по умолчанию у клиента

- `HOST=localhost:8080`
- `GRPC_SERVER_ADDRESS` по умолчанию пустой, поэтому его лучше задавать явно

### Пример запуска клиента через ENV

#### Linux / macOS

```bash
export HOST='localhost:8080'
export GRPC_SERVER_ADDRESS='localhost:5051'

./gophkeeper-client
```

#### Windows PowerShell

```powershell
$env:HOST='localhost:8080'
$env:GRPC_SERVER_ADDRESS='localhost:5051'

./gophkeeper-client.exe
```

### Пример JSON-конфига клиента

`config.client.json`

```json
{
  "host": "localhost:8080",
  "grpc_server_address": "localhost:5051"
}
```

Запуск:

```bash
./gophkeeper-client -c ./config.client.json
```

или:

```bash
CONFIG=./config.client.json ./gophkeeper-client
```

---

## Рекомендуемый порядок локального запуска

Сначала поднять сервер:

```bash
./gophkeeper-server -d 'postgres://postgres:postgres@localhost:5432/gophkeeper?sslmode=disable' -g ':5051' -k 'BASE64_КЛЮЧ_ИЗ_32_БАЙТ' -a 'super-secret-auth-key' -df 24 -t 24
```

Затем поднять клиент:

```bash
./gophkeeper-client -a 'localhost:8080' -g 'localhost:5051'
```

---

## Swagger / OpenAPI

В клиентском HTTP-приложении уже есть swagger-аннотации в `main.go`:

- title: `GophKeeper API`
- version: `1.0`
- host: `localhost:8080`
- base path: `/api`
- security scheme: `Authorization: Bearer {token}`

Это значит, что проект уже подготовлен для генерации Swagger-документации через `swaggo/swag`.

### Установка генератора

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### Генерация swagger-документации

Если main клиента лежит в `./cmd/client/main.go`:

```bash
swag init -g ./cmd/client/main.go -o ./docs
```

Если main клиента лежит в текущей директории:

```bash
swag init -g ./main.go -o ./docs
```

После этого обычно появляются файлы вроде:

- `docs/docs.go`
- `docs/swagger.json`
- `docs/swagger.yaml`

### Публикация Swagger UI

Чтобы Swagger UI был доступен из приложения, обычно делают так:

1. импортируют сгенерированный пакет `docs`;
2. подключают HTTP-handler для Swagger UI;
3. регистрируют маршрут, например `/swagger/*`.

Типовой вариант для Go:

```go
import (
    _ "your-module/docs"
    httpSwagger "github.com/swaggo/http-swagger"
)
```

И затем добавить маршрут, например:

```go
router.Get("/swagger/*", httpSwagger.WrapHandler)
```

### Где открывать Swagger UI

Если swagger handler подключен, документация обычно доступна по одному из путей:

```text
http://localhost:8080/swagger/index.html
http://localhost:8080/swagger/
```

### Авторизация в Swagger

В аннотациях уже задан `ApiKeyAuth` через заголовок `Authorization`, поэтому в Swagger UI токен обычно передают так:

```text
Bearer <your_token>
```

### Важно

В загруженных файлах видны swagger-аннотации, но сам код публикации Swagger UI не показан. Поэтому:

- **генерация docs уже возможна**;
- **маршрут Swagger UI нужно проверить в handler/router части проекта**;
- если маршрут еще не подключен, его нужно добавить вручную.

---

## Остановка приложений

И сервер, и клиент корректно обрабатывают:

- `SIGINT`
- `SIGTERM`

На `SIGQUIT` выводится dump горутин в `stderr`.
