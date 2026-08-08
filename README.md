# CodeBattle MVP

Рабочий MVP сервиса дуэлей по решению задач на Go. Два пользователя находят друг друга в общем lobby, отправляют приглашение, видят одну задачу и код соперника, запускают решение в изолированном judge и продолжают серию после готовности обоих.

## Что уже работает

- минимальная регистрация и вход по `username`/паролю;
- Argon2id, opaque HttpOnly-сессии и presence в PostgreSQL;
- список всех пользователей, поиск, cursor-пагинация и статусы;
- приглашения с TTL 30 секунд и атомарное создание матча;
- 20 версионируемых Go-задач с public/hidden-тестами;
- Monaco Editor, локальный Go-autocomplete и синхронизация снимков кода с ревизиями;
- очередь отправок в PostgreSQL и отдельный judge-worker;
- постоянный прогреваемый Go build cache для быстрых повторных компиляций;
- изолированные compile/runtime Docker-контейнеры без сети;
- результаты `accepted`, `wrong_answer`, `compile_error`, `runtime_error` и `time_limit`;
- счёт серии, победитель раунда, ready-flow и выбор задач без повторов;
- пауза при разрыве соединения и восстановление в течение 60 секунд;
- тёмная/светлая темы Mantine, синхронизированные с Monaco;
- health/readiness endpoints, JSON-логи, security headers, Origin-guard и базовые rate limits.

В упрощённой версии Redis и WebSocket намеренно не используются: PostgreSQL хранит всё авторитетное состояние, очередь judge и code snapshots, а клиенты получают обновления коротким polling. Для одного API-инстанса и первых пользователей этого достаточно; интерфейсы можно перевести на WebSocket без изменения модели матча.

## Быстрый запуск

Требуется Docker Desktop с Linux containers.

```bash
docker compose -f docker-compose.yml -f docker-compose.local.yml up -d --build
```

После запуска приложение доступно на [http://localhost:8088](http://localhost:8088). Состояние контейнеров:

```bash
docker compose -f docker-compose.yml -f docker-compose.local.yml ps
```

Остановка:

```bash
docker compose -f docker-compose.yml -f docker-compose.local.yml down
```

PostgreSQL-данные сохраняются в named volume. Чтобы удалить их вместе со стеком, добавьте `-v` к команде `down` только осознанно.

## Полный smoke-сценарий

При работающем стеке команда создаёт двух уникальных пользователей, проводит invite-flow, проверяет снимок кода, получает безопасный `wrong_answer` для неверного решения, затем запускает эталонное решение через настоящий judge, проверяет счёт и начинает второй раунд:

```bash
go run ./cmd/smoke-duel
```

Каталог задач можно проверить отдельно:

```bash
go run ./cmd/problem-seed -validate-only
```

## Локальная разработка

Требования: Go 1.26+, Node.js 22.20+, npm 10.9+ и PostgreSQL 17. Проще оставить PostgreSQL из Compose запущенным или задать собственный `DATABASE_URL`.

```bash
npm ci --prefix apps/web
go run ./cmd/api
npm run dev
```

Vite открывается на `http://localhost:5173` и проксирует API на `http://localhost:8080`. Значения окружения перечислены в `.env.example`. API самостоятельно применяет ещё не выполненные миграции; `problem-seed` загружает неизменяемые версии задач.

## Проверки

```bash
go test ./...
go vet ./...
npm run typecheck
npm test
npm run build
docker compose config --quiet
docker compose -f docker-compose.yml -f docker-compose.local.yml config --quiet
```

Те же проверки запускаются workflow `.github/workflows/ci.yml`.

## Judge и границы безопасности

Только judge-worker получает Docker socket. API не видит hidden tests и не имеет доступа к Docker daemon. Компиляция и тесты выполняются в разных контейнерах; runtime получает только test binary. Контейнеры запускаются без сети и Linux capabilities, от non-root пользователя, с read-only rootfs, `no-new-privileges`, PID/CPU/memory/output limits. Внутренние пути и содержимое hidden tests не возвращаются клиенту.

Это практичная изоляция для MVP на отдельной Linux VM, но Docker socket даёт worker высокий уровень доверия. Перед многоарендным публичным запуском worker следует вынести на отдельный хост или заменить специализированной sandbox-средой.

## Документация

- [Полная техническая спецификация](docs/codebattle-mvp-spec.md)
- [Этапы и фактический статус реализации](docs/implementation-roadmap.md)
- [Правила миграций](db/migrations/README.md)
- [Production-развёртывание, backup и обновление](docs/operations.md)
- [Развёртывание через Dokploy](docs/dokploy.md)
