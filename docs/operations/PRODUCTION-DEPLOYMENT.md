# Production deployment runbook

## Required secrets
Set DATABASE_URL, POSTGRES_PASSWORD and ADMIN_API_TOKEN outside source control. ADMIN_API_TOKEN must be a long random secret. Set NEXT_PUBLIC_API_BASE_URL to the public HTTPS API origin and API_INTERNAL_BASE_URL to the private API service origin.

## Pre-deploy
1. Take and verify a PostgreSQL backup.
2. Record the currently deployed image/commit for rollback.
3. Run database migrations as a one-off release step before replacing application containers.
4. Confirm /healthz and /readyz return success.
5. Never expose PostgreSQL, Redis or OpenSearch ports to the public Internet.

## Database migrations
Migrations are ordered and recorded in schema_migrations. Each unapplied migration executes in a transaction. Do not edit an already-applied migration in production; add a new forward migration.

## Backup
Example from the Docker host:
```sh
docker compose exec -T postgres pg_dump -U sodienthoai -Fc sodienthoai > backup-$(date +%Y%m%d-%H%M%S).dump
```
Store backups outside the application host and periodically test restore into a disposable database.

## Deploy
Pull/build the exact reviewed commit, run migrations, start services, then verify readiness and a known lookup. Keep the previous image available until smoke tests pass.

## Rollback
For application-only failures, restore the previous API/web image. Database rollback is not automatic: prefer forward-compatible migrations and corrective forward migrations. Restore a database backup only for a confirmed destructive data incident.

## Network and TLS
Terminate TLS at a reverse proxy or managed load balancer. Redirect HTTP to HTTPS. Expose only 80/443 publicly; keep API dependencies on a private network. Apply firewall rules at the host/cloud layer.

## Monitoring
Alert on API readiness failures, 5xx rate, latency, database capacity, Redis/OpenSearch health, disk usage and backup failures. Centralize application and proxy logs with retention.

## Release gate
A release is eligible only when exact-head CI is green for API, database, web, containers and integration tests, migrations succeed on a staging/fresh database, backup is verified, secrets are present and post-deploy smoke tests pass.
