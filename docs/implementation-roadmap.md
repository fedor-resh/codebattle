# CodeBattle: этапы и статус реализации

MVP реализован последовательными продуктовыми срезами. По просьбе не переусложнять первый релиз Redis и WebSocket отложены: авторитетное состояние, сессии, presence, code snapshots и очередь judge находятся в PostgreSQL, обновления интерфейса доставляются polling-запросами.

## Этап 1. Фундамент — завершён

- [x] Monorepo, общие npm-команды и конфигурация Go.
- [x] React, TypeScript, Vite, MantineProvider, тема и AppShell.
- [x] Monaco Editor с lazy-loaded экраном матча.
- [x] Go API, middleware, health checks и graceful shutdown.
- [x] PostgreSQL migrations и автоматическое применение при старте API.
- [x] Compose для PostgreSQL, seed, API, judge-worker и Caddy gateway.
- [x] Docker images собраны и полный стек запущен локально.
- [x] CI для tests, vet, typecheck, frontend build и Compose validation.

## Этап 2. Аккаунты и lobby — завершён

- [x] Минимальная регистрация и вход по username/password.
- [x] Проверка username, Argon2id и opaque HttpOnly-cookie.
- [x] Сессии и presence timestamp в PostgreSQL.
- [x] Список пользователей, поиск, cursor-пагинация и нужная сортировка.
- [x] Статусы available/busy/offline без опоры только на цвет.
- [x] Responsive Mantine UI и сохранение light/dark темы.

## Этап 3. Приглашения и комнаты — завершён

- [x] TTL приглашения 30 секунд и явные конфликтные состояния.
- [x] Одно активное приглашение на пользователя.
- [x] Create/state/accept/decline REST-команды.
- [x] Row locks и атомарное создание единственного матча.
- [x] Incoming modal, countdown и автоматический переход обоих игроков.
- [x] Выход из серии с подтверждением.

## Этап 4. Задачи и judge — завершён

- [x] YAML-схема, строгий loader, AST-проверка и content hash.
- [x] 20 задач со starter, statement, public/hidden tests и reference solution.
- [x] Идемпотентный seed неизменяемых problem versions.
- [x] Очередь отправок в PostgreSQL с `FOR UPDATE SKIP LOCKED`.
- [x] Лимит одной отправки в две секунды и до трёх незавершённых.
- [x] Раздельные compile/runtime контейнеры с sandbox-ограничениями.
- [x] Очистка compiler output и закрытый ответ для hidden test.
- [x] Транзакционное определение победителя и однократное изменение счёта.

## Этап 5. Полная дуэль — завершён

- [x] Снимки полного текста с debounce 150 мс и монотонной revision.
- [x] Авторитетный snapshot в PostgreSQL и защита от устаревшей revision.
- [x] Автоматическое отображение кода соперника только для чтения.
- [x] Статусы judge и панель результата в интерфейсе.
- [x] Блокировка раунда после победы и показ победного кода.
- [x] Ready-flow, общий счёт и бесконечная серия.
- [x] Задачи без повторов до исчерпания пула; граница циклов без повтора.
- [x] Пауза/восстановление 60 секунд и завершение покинутой серии.
- [x] Двухпользовательский API smoke и браузерная проверка двух сессий.

## Этап 6. Готовность MVP — завершён

- [x] Security headers и проверка Origin для изменяющих запросов.
- [x] Rate limits регистрации, входа и приглашений; доменные лимиты judge.
- [x] Очистка просроченных сессий и приглашений.
- [x] API readiness проверяет PostgreSQL; worker readiness проверяет PostgreSQL и Docker.
- [x] Структурированные JSON-логи с request/submission/match identifiers.
- [x] Unit, integration smoke, production build и реальная Docker-проверка judge.
- [x] Browser QA invite-flow, match screen, opponent pane, theme switch и leave-flow.
- [x] Документация запуска, проверок, архитектурных компромиссов и безопасности.
- [x] Production Compose с Caddy/ACME TLS и backup/restore runbook.

## Следующие улучшения после MVP

Это не блокеры текущего релиза, а точки роста при появлении реальной нагрузки:

- WebSocket вместо polling для меньшей задержки и числа запросов;
- Redis или брокер для нескольких API/worker-инстансов;
- Prometheus/OpenTelemetry и нагрузочный профиль 200 пользователей/100 комнат;
- отдельный sandbox-host для judge;
- полноценный автоматизированный Playwright E2E в CI.
