# V40 staging verification

V40 remains staging-only until every check below passes.

## Database
- Back up the staging PostgreSQL database.
- Apply migrations 000018, 000019 and 000020 with the normal migration runner.
- Confirm schema_migrations records all three versions.
- Confirm API readiness remains healthy after migration.
- Do not apply these migrations to production during V40 staging validation.

## Account parity
- Authenticated user can set trusted/blocked preference.
- Another user cannot read or mutate that preference.
- Authenticated user can list and delete their own lookup history.
- Appeal requires authentication, valid reason and 10-4000 character statement.
- Duplicate pending appeal returns conflict.

## Business
- Authenticated user can create an unverified business.
- Unverified business is not returned by public business API or business sitemap.
- Only an owner/manager record for the business can request verification.
- Verification evidence reference is never returned by public business API.
- Admin verification approval marks the business and requesting ownership verified.
- Review creation is rejected until the business is verified.
- A branch review is rejected if the branch belongs to another business.
- Public profile aggregates only approved reviews.

## Moderation and audit
- Admin queue endpoints reject requests without the configured admin bearer token.
- Appeal, verification and review moderation endpoints reject unauthorized requests.
- Successful moderation creates an admin_audit_log event.
- Repeating a completed moderation transition is rejected.

## SEO
- Phone sitemap behavior remains unchanged.
- Business sitemap contains verified businesses only.
- No verification evidence or account-owned data appears in sitemap/public responses.

## Regression
- Run the full repository CI suite.
- Smoke-test /healthz and /readyz.
- Smoke-test existing lookup, report, claim, follow, comments and notification flows.
- Confirm staging web uses staging API routing.
- Keep sodienthoai.com and production compose unchanged.


## V42 hardening gates
- Exercise business verification through the real HTTP account proxy path, not only direct database fixtures.
- Run concurrent duplicate verification and review requests; exactly one active request/review may survive.
- Reject verification from unrelated, rejected or revoked ownership and prove no business state mutation occurs.
- Reject replay/invalid moderation transitions, including stale requests after ownership revocation or business suspension.
- Check public, account and admin response shapes for verification evidence leakage; evidence_ref remains private.
- Validate pending/rejected/suspended businesses never enter public business API, JSON-LD or business sitemap.
- Verify Business Center maps expected API errors to user-facing messages and does not offer invalid verification actions.
- Run final full CI on the exact V42 head before merge.
