# V40 — Feature Parity Foundation

## Goal

Bring forward the useful business capabilities from the legacy WordPress implementation into the Go/PostgreSQL/Next.js platform without reproducing WordPress-specific architecture.

Production cutover is explicitly out of scope. V40 is built and validated on staging first.

## Source audit

The legacy implementation provides useful domain behavior in these areas:

- phone lookup and lookup history
- account preferences, blocked/trusted numbers
- user reports, abuse controls and reputation/risk recalculation
- appeals against phone labels
- business profiles, branches, ownership and document verification
- business reviews and moderation
- numbering-plan management/import
- moderation queues, anomaly review and audit trail
- background jobs/system statistics
- extension reputation API (single and batch)
- configurable advertising slots
- SEO eligibility rules for phone and verified-business pages

## Product decisions

### P0 — implement/complete first

1. Phone lookup remains the core public workflow.
2. Account dashboard exposes lookup history and trusted/blocked numbers.
3. Reports feed reputation/risk through abuse-resistant aggregation.
4. Appeals are first-class records with pending/accepted/rejected workflow.
5. Business profiles support branches and phone links.
6. Business ownership/verification uses a reviewable verification workflow.
7. Reviews require moderation state.
8. Admin has queues for reports, appeals, reviews and business verification.
9. All moderation mutations write an audit event.
10. SEO indexing is eligibility-based; an unknown/thin phone page is not automatically indexable.

### P1 — implement after P0

- numbering-plan administration/import/export
- anomaly detection views
- system statistics and operational job visibility
- extension compatibility endpoints for single/batch reputation
- legacy URL compatibility/redirect map

### P2 — keep optional/decoupled

- advertising slots
- legacy demo-content behavior

Ads must not be coupled to lookup/reputation domain logic.

## Architecture decisions

- PostgreSQL is the source of truth.
- Redis is for cache/rate limiting/ephemeral coordination, not authoritative records.
- OpenSearch remains a search/index projection.
- Risk/reputation recalculation is an explicit job/event workflow, not a WordPress-style periodic page-runtime cron.
- Verification documents are private objects; public/business APIs must never expose storage paths.
- Moderation endpoints require authenticated admin authorization.
- User-facing report/appeal/review writes are rate limited and auditable.
- Keep API compatibility adapters separate from core domain services.

## SEO policy

A phone URL may exist for lookup UX without being eligible for search indexing.

Initial parity rule to preserve from the legacy behavior:

- index a phone when it has meaningful trust/risk/community evidence; do not sitemap every synthetically possible number
- index a business only when verified
- sitemap generation must use stable keyset/sharded traversal and eligibility filters
- thin/unknown pages should be noindex until eligibility is reached

Exact thresholds remain configurable and must be covered by tests before production cutover.

## Compatibility routes

Legacy account paths to preserve through route aliases or redirects at cutover:

- /dang-nhap
- /dang-ky
- /quen-mat-khau
- /tai-khoan
- /tai-khoan/chan-so
- /tai-khoan/thoat

Extension compatibility should support the semantic equivalent of:

- single phone reputation
- batch reputation
- capabilities
- health/live/ready

The new native API remains canonical.

## Acceptance gates

V40 is not production-ready until:

- database migrations are reversible and pass CI
- P0 API integration tests pass
- authorization tests cover every moderation mutation
- report/review/appeal rate-limit tests pass
- business verification does not leak private document locations
- SEO tests prove unknown/thin phone pages are excluded from sitemap/index eligibility
- staging migrations succeed without damaging existing V39 data
- staging API readyz and web smoke tests pass
- production domain remains unchanged
