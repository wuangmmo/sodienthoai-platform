# Local development

## Infrastructure

Copy `.env.example` to `.env`, then start the backing services:

```sh
docker compose up -d postgres redis opensearch
```

Expected local ports:

- PostgreSQL: 5432
- Redis: 6379
- OpenSearch: 9200
- Go API: 8080
- Next.js web: 3000

## Health contract

The API exposes `GET /healthz` for process-level liveness. Dependency readiness will be added separately so a temporary database/search outage does not incorrectly kill the application process.

## Production rule

Local Docker Compose is a development convenience, not the production topology. Production credentials must not use the development values in this repository.
