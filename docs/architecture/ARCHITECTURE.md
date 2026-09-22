# SoDienThoai Platform Architecture V1

## Principles

1. PostgreSQL is the source of truth.
2. OpenSearch is a rebuildable search index, never the canonical database.
3. Redis is an optimization layer, never authoritative storage.
4. The public web is SEO-aware by design.
5. A phone record existing does not imply that its URL should be indexed.
6. Store canonical phone identities in E.164 form.
7. Start as a modular monolith and extract services only when operational evidence justifies it.

## Request path

Cloudflare -> Next.js -> Go API -> Redis / PostgreSQL / OpenSearch

Asynchronous workers will handle search indexing, quality scoring, sitemap eligibility and later analytics/event processing.

## SEO lifecycle

discovered -> noindex -> indexable -> indexed -> review

Only sufficiently useful records become indexable and sitemap-eligible. Empty or near-duplicate phone pages must not be mass-indexed.

## Scale path

V1: PostgreSQL + Redis + OpenSearch, single-region application.
V2: workers, queues, read replicas and dedicated search cluster.
V3: partitioning, independent high-load services and analytics store.
V4: multi-region reads and regionalized data/search infrastructure.
V5: distributed global architecture based on measured load and data-residency needs.
