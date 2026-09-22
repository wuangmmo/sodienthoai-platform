# Local development

## Start the platform

Copy the example environment file and start the local stack:

```sh
cp .env.example .env
docker compose up --build
```

Services:
- Web: http://localhost:3000
- API: http://localhost:8080
- PostgreSQL: localhost:5432
- Redis: localhost:6379
- OpenSearch: http://localhost:9200

## Health endpoints

- `GET /healthz` is process liveness only.
- `GET /readyz` checks PostgreSQL, Redis, and OpenSearch and returns 503 until all required dependencies are ready.

## Database migrations

Migrations are explicit and are never applied implicitly by the API process.

```sh
cd apps/api
DATABASE_URL='postgres://sodienthoai:sodienthoai@localhost:5432/sodienthoai?sslmode=disable' \
MIGRATIONS_DIR='../../database/migrations' go run ./cmd/migrate
```

## SEO quality evaluation

```sh
cd apps/api
DATABASE_URL='postgres://sodienthoai:sodienthoai@localhost:5432/sodienthoai?sslmode=disable' go run ./cmd/seo-evaluate
```

Only records meeting the quality rules become `indexable`; database existence alone never makes a phone page indexable.

## Rebuild OpenSearch

PostgreSQL remains the source of truth. OpenSearch can be rebuilt:

```sh
cd apps/api
DATABASE_URL='postgres://sodienthoai:sodienthoai@localhost:5432/sodienthoai?sslmode=disable' \
OPENSEARCH_URL='http://localhost:9200' REINDEX_BATCH_SIZE=1000 go run ./cmd/reindex
```

## Phone lookup

Vietnamese national input must carry explicit country context:

```text
GET /v1/phone/0705899899?country=VN
```

Canonical international input works directly:

```text
GET /v1/phone/+84705899899
```

The web application redirects successful national-format lookups to the canonical E.164 phone URL.
