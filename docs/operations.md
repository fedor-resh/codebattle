# CodeBattle MVP: эксплуатация

## Развёртывание на Linux VM

Нужны Docker Engine, Compose plugin, домен с A/AAAA-записью на VM и открытые порты 80/443. Создайте `.env` вне системы контроля версий:

```dotenv
DOMAIN=battle.example.com
POSTGRES_PASSWORD=replace-with-a-long-random-value
POSTGRES_USER=codebattle
POSTGRES_DB=codebattle
```

Запуск с автоматическим ACME TLS от Caddy:

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build
docker compose -f docker-compose.yml -f docker-compose.prod.yml ps
curl -fsS https://battle.example.com/health/ready
```

API в production ставит cookie `Secure`; изменяющие запросы принимаются только с `https://$DOMAIN`.

## Резервное копирование

Создайте каталог, доступный только оператору, и снимите PostgreSQL custom-format backup:

```bash
mkdir -p backups
docker compose exec -T postgres sh -c 'pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Fc -f /tmp/codebattle.dump'
docker compose cp postgres:/tmp/codebattle.dump backups/codebattle-$(date +%F-%H%M).dump
docker compose exec -T postgres rm -f /tmp/codebattle.dump
```

Скопируйте backup за пределы VM и настройте retention. Каталоги judge временные и в backup не входят; problem bundles находятся в PostgreSQL.

## Проверка восстановления

Восстановление выполняйте на отдельной тестовой базе, не поверх работающего production:

```bash
docker compose cp backups/codebattle-YYYY-MM-DD-HHMM.dump postgres:/tmp/restore.dump
docker compose exec -T postgres createdb -U "$POSTGRES_USER" codebattle_restore
docker compose exec -T postgres pg_restore -U "$POSTGRES_USER" -d codebattle_restore /tmp/restore.dump
docker compose exec -T postgres psql -U "$POSTGRES_USER" -d codebattle_restore -c 'select count(*) from problem_versions;'
docker compose exec -T postgres dropdb -U "$POSTGRES_USER" codebattle_restore
docker compose exec -T postgres rm -f /tmp/restore.dump
```

## Обновление

1. Снять backup и проверить свободное место.
2. Получить новую ревизию репозитория.
3. Выполнить `docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build`.
4. Проверить `ps`, `/health/ready`, логи API/worker и запустить `go run ./cmd/smoke-duel -base-url https://$DOMAIN/api/v1` с доверенного хоста.

Миграции применяются API один раз и записываются в `schema_migrations`. Problem seed идемпотентен: существующая версия с тем же content hash повторно не создаётся. Версию Go и `JUDGE_IMAGE` обновляйте вместе; после обновления обязательно прогоняйте каталог задач и smoke duel.

## Диагностика

```bash
docker compose ps
docker compose logs --since=15m api judge-worker
docker compose exec postgres pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB"
docker system df
```

Не публикуйте Docker socket или PostgreSQL port наружу. В логах не должно быть паролей, session tokens, исходников hidden tests или полного пользовательского кода.
