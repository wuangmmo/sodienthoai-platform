# SEO & Indexing V5

The phone corpus and the Google-indexable corpus are deliberately separate.

- Newly discovered/imported numbers remain noindex.
- A number becomes indexable only after identity/quality eligibility.
- Sitemap partitions target 40,000 URLs, below the protocol URL limit.
- Canonical phone paths use the normalized E.164 representation.
- noindex pages remain useful to users and can collect reports/claims without polluting sitemap output.
- Indexing is promoted in controlled batches; database presence never implies search-engine eligibility.

## Index lifecycle
SEO status changes are queued in `seo_index_queue`. This decouples page eligibility from external indexing/search synchronization and allows retries without blocking public lookups.

## Crawl controls
Unknown, imported, disputed, and low-quality records remain `noindex,follow`. Only qualified `indexable` or already `indexed` records enter segmented sitemaps. Sitemap pages are capped at 40,000 URLs to stay below protocol limits and leave operational headroom.

## Canonical and structured data
Phone pages canonicalize to normalized E.164 URLs. Qualified pages expose Open Graph metadata and minimal JSON-LD based on the verified public identity.
