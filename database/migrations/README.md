# Database migrations

Migrations are immutable once released.

Files use sequential IDs:

- `000001_core.up.sql`
- `000001_core.down.sql`

Production migrations must run as an explicit deployment step before new application instances become ready. The application process does not silently mutate the production schema at startup.

PostgreSQL remains the canonical source of truth. Redis and OpenSearch data must be rebuildable from canonical data/events.
