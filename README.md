# SoDienThoai Platform

Phone-number intelligence, verification, reputation, public Internet footprint and private contact intelligence platform.

## Runtime
Go API, PostgreSQL, Redis, OpenSearch, Next.js web/admin and Docker Compose.

## Production gate
Do not deploy production until the current main commit has all five CI jobs green, a fresh database backup has been verified, required environment variables are configured locally, migrations have been reviewed, and staging smoke tests pass.

Unknown numbers remain discovery/noindex until quality rules promote them. Private contact labels are never automatically promoted to public identity.
