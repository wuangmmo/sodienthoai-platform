# Deployment and rollback

Deploy only an exact commit that has passed all five CI groups. Create and verify a fresh database backup first.

The deployment script validates Compose, builds the exact checked-out code, starts services without deleting persistent volumes, and requires API readiness before reporting success.

Rollback is code-only: choose a known-good commit and redeploy it. It deliberately does not run down -v, delete volumes, or reverse database migrations. If a migration is not backward compatible, stop and use the reviewed migration-specific recovery plan rather than forcing an older binary against a newer schema.

After every staging or production deployment, verify readiness, public phone lookup, an unknown-number discovery response, admin login, a protected admin page, sitemap output, and one user-session mutation.
