# V36 Production Release Gate

V36 is the final verification gate before a production deployment.

## Required checks

- API unit tests and static checks pass.
- Database migrations apply cleanly to a fresh PostgreSQL instance.
- Web build succeeds with committed dependency locks.
- API and web containers build successfully.
- Integration tests pass against PostgreSQL, Redis and OpenSearch.
- /healthz and /readyz are healthy before traffic is switched.
- Production secrets are supplied externally and are not committed.
- Unknown or thin phone-number records remain noindex.
- Admin and user session secrets are distinct production-grade values.
- Reporter hashing secret and admin API token are rotated for production.
- OpenSearch remains rebuildable from PostgreSQL.
- A database backup is captured before applying production migrations.

## Release procedure

1. Back up PostgreSQL.
2. Build immutable API and web images from the approved commit.
3. Apply migrations once.
4. Start dependencies and application services.
5. Verify health and readiness.
6. Smoke-test lookup, profile, report, claim, community and admin flows.
7. Verify robots and sitemap behavior.
8. Switch traffic only after all checks pass.

## Rollback

- Switch traffic back to the previous immutable application images.
- Do not automatically roll back destructive database migrations.
- Restore the database backup only when a migration/data incident requires it.
- Rebuild OpenSearch from PostgreSQL when search state is inconsistent.

A release must not be marked production-ready while any required check is failing.
