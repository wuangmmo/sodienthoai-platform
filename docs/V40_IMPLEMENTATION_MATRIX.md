# V40 implementation matrix

| Capability | Existing platform | V40 action | Priority |
|---|---|---|---|
| Phone lookup/profile | implemented | retain and harden | P0 |
| Internet footprint | implemented | retain | P0 |
| Reports + duplicate protection | implemented | extend moderation/reputation events | P0 |
| Phone claims | implemented | retain; separate from business verification | P0 |
| Follow + notifications | implemented | retain | P0 |
| Community comments | implemented | treat as community discussion, not business review | P0 |
| Account lookup history | aggregate lookup events only | add per-user history | P0 |
| Trusted/blocked numbers | absent | add user preferences | P0 |
| Appeals | absent | add workflow + admin moderation | P0 |
| Business profiles | absent | add businesses/branches/phone links | P0 |
| Business verification | absent | add private verification workflow | P0 |
| Business reviews | absent | add moderated reviews | P0 |
| SEO eligibility | implemented via seo_status | tighten evidence policy and tests | P0 |
| Admin audit | implemented | require on all new moderation writes | P0 |
| Numbering plans | absent | add after P0 workflows | P1 |
| Extension reputation compatibility | absent | adapter after native API is stable | P1 |
| Ads | absent | optional decoupled module | P2 |

## Important distinctions

- phone_claims proves/control-disputes a phone identity. It is not a substitute for a verified business entity.
- phone_comments are community discussion. They are not a substitute for a structured business review/rating.
- phone_lookup_events are anonymous/global analytics. user_lookup_history is an account-owned convenience feature and must be independently deletable.
- Trusted/blocked is a user's private preference and must not alter the global reputation score.
- An appeal requests review of a platform label/status; it must not directly mutate reputation without moderation.

## V40 migration sequence

- 000018: account lookup history, trusted/blocked preferences, appeals
- 000019: businesses, branches, phone links, verification requests
- 000020: business reviews and moderation indexes
- subsequent migrations only when API/tests for the preceding slice are stable
