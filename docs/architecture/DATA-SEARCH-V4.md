# Data & Search V4

V4 is designed for large phone-number datasets.

## Import
CSV columns: country_code, calling_code, national_number, e164, number_type.
The import command streams the input and writes bounded batches instead of loading the full file into memory. Imported numbers default to noindex.

## Search
The public API prefers OpenSearch and falls back to PostgreSQL. PostgreSQL pagination uses a cursor over search_count/e164 rather than OFFSET for deep result sets.

## Indexing
The reindex command streams PostgreSQL rows into OpenSearch bulk requests and swaps the phone_numbers alias after a successful rebuild. Primary public identity names are indexed for name search.

## SEO separation
Importing or discovering a number never makes it indexable automatically. SEO eligibility remains controlled by verification, identity confidence and data-quality rules.
