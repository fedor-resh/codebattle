# Развёртывание CodeBattle через Dokploy

Используйте сервис типа **Docker Compose**, а не Docker Stack.

## General

- Repository/branch: production-репозиторий и ветка `main`.
- Compose Path: `./docker-compose.yml`.
- Isolated Deployments: включено.
- Custom Compose Command: пусто, используется команда Dokploy по умолчанию.
- Auto Deploy: включить после первого успешного ручного deployment.

Базовый Compose объявляет порт `80` без фиксированного host port: Docker автоматически выбирает гарантированно свободный динамический порт. Dokploy подключает Traefik к сервису `gateway` на внутреннем порту 80. Файл `docker-compose.local.yml` предназначен только для локальной разработки и фиксирует локальный адрес `8088:80`; в Dokploy он не используется. Файл `docker-compose.prod.yml` поднимает собственный Caddy TLS и также не используется в Dokploy, поскольку TLS завершает Traefik.

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

Выделенный Docker host port можно увидеть в deployment preview/container details или командой `docker compose port gateway 80`. Он предназначен для прямой диагностики через `http://SERVER_IP:PORT`; в настройке домена по-прежнему указывается **Container Port 80**, а не динамический host port.

## Persistent data и backup

PostgreSQL использует named volume `postgres-data`. В Dokploy Volume Backups выберите Compose service `postgres`, этот volume и S3 destination. Judge source/binary volumes содержат только временные файлы, а `judge-cache` — восстанавливаемый Go build cache; их в backup не включают.

`judge-worker` должен сохранить единственный bind mount Docker socket `/var/run/docker.sock`; он нужен для запуска sandbox-контейнеров. Не добавляйте этот mount другим сервисам.

## Проверка

У `postgres`, `api`, `judge-worker` и `gateway` должен быть статус healthy. Одноразовый `problem-seed` должен завершиться с кодом 0.

```bash
curl -fsS https://battle.example.com/health/ready
go run ./cmd/smoke-duel -base-url https://battle.example.com/api/v1
```
