# Развёртывание CodeBattle через Dokploy

Используйте сервис типа **Docker Compose**, а не Docker Stack.

## General

- Repository/branch: production-репозиторий и ветка `main`.
- Compose Path: `./docker-compose.yml`.
- Isolated Deployments: включено.
- Custom Compose Command: пусто, используется команда Dokploy по умолчанию.
- Auto Deploy: включить после первого успешного ручного deployment.

Базовый Compose не публикует host ports. Dokploy подключает Traefik к сервису `gateway` на внутреннем порту 80. Файл `docker-compose.local.yml` предназначен только для локальной разработки и в Dokploy не используется. Файл `docker-compose.prod.yml` поднимает собственный Caddy TLS и также не используется в Dokploy, поскольку TLS завершает Traefik.

## Environment

```dotenv
APP_ENV=production
POSTGRES_DB=codebattle
POSTGRES_USER=codebattle
POSTGRES_PASSWORD=<64-символьный случайный hex>
PUBLIC_ORIGINS=https://battle.example.com
```

Пароль можно сгенерировать командой `openssl rand -hex 32`. Используйте значение без URL-reserved символов, потому что Compose подставляет пароль в `DATABASE_URL`.

## Domain

- Host: `battle.example.com`.
- Service Name: `gateway`.
- Container Port: `80`.
- Path/Internal Path: `/`.
- Strip Path: выключено.
- HTTPS и Let's Encrypt: включены.

После добавления или изменения домена выполните Redeploy. DNS A/AAAA-запись должна указывать на Dokploy server; TCP-порты 80 и 443 должны быть доступны извне.

## Persistent data и backup

PostgreSQL использует named volume `postgres-data`. В Dokploy Volume Backups выберите Compose service `postgres`, этот volume и S3 destination. Judge volumes содержат только временные файлы и в backup не включаются.

`judge-worker` должен сохранить единственный bind mount Docker socket `/var/run/docker.sock`; он нужен для запуска sandbox-контейнеров. Не добавляйте этот mount другим сервисам.

## Проверка

У `postgres`, `api`, `judge-worker` и `gateway` должен быть статус healthy. Одноразовый `problem-seed` должен завершиться с кодом 0.

```bash
curl -fsS https://battle.example.com/health/ready
go run ./cmd/smoke-duel -base-url https://battle.example.com/api/v1
```
