# Backup and restore runbook

Production changes require a fresh PostgreSQL backup first.

## Backup
Run `scripts/backup-production.sh` from the repository root with the production environment loaded. The script uses pg_dump from the production PostgreSQL container, compresses the dump, verifies gzip integrity, rejects empty output, and restricts file permissions.

Keep at least one verified copy outside the application host.

## Restore verification
Never test a restore by overwriting the live production database. Restore the latest dump into an isolated PostgreSQL instance, run migrations/readiness checks, and verify representative phone, identity, report, contact and community records.

The included `restore-production.sh` intentionally refuses an automatic live restore. Production restore remains a deliberate operator action until an isolated restore-test workflow is added.

## Before deploy
Record the main commit, verify all five CI groups are green, create and verify a backup, confirm required environment variables are configured without printing them, then deploy. Do not delete Docker volumes during rollback.
