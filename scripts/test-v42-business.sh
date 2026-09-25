#!/usr/bin/env bash
set -euo pipefail
API="${API_BASE_URL:-http://localhost:18080}"
DB="${DATABASE_URL:?DATABASE_URL is required}"
TOKEN="${ADMIN_API_TOKEN:?ADMIN_API_TOKEN is required}"
PSQL=(psql "$DB" -v ON_ERROR_STOP=1 -At)

sql(){ "${PSQL[@]}" -c "$1"; }
code(){ curl -sS -o "$2" -w '%{http_code}' "${@:3}"; }

# Isolated V42 fixtures.
sql "DELETE FROM businesses WHERE slug LIKE 'v42-ci-%';"
sql "INSERT INTO user_accounts(subject_key) VALUES('v42-owner'),('v42-other'),('v42-reviewer') ON CONFLICT(subject_key) DO NOTHING;"
sql "INSERT INTO businesses(slug,legal_name,display_name,verification_status,created_by) SELECT 'v42-ci-concurrent','V42 Concurrent','V42 Concurrent','unverified',id FROM user_accounts WHERE subject_key='v42-owner';"
BID=$(sql "SELECT id FROM businesses WHERE slug='v42-ci-concurrent';")
OWNER=$(sql "SELECT id FROM user_accounts WHERE subject_key='v42-owner';")
sql "INSERT INTO business_ownerships(business_id,user_id,role,status) VALUES('$BID','$OWNER','owner','pending');"

# Two simultaneous verification requests: exactly one accepted.
for n in 1 2; do
  (curl -sS -o "/tmp/v42-verify-$n.json" -w '%{http_code}' -H 'X-User-Subject: v42-owner' -H 'Content-Type: application/json' -d '{"method":"manual","statement":"V42 concurrency","evidence_ref":"private://v42-evidence"}' "$API/v1/businesses/$BID/verification-requests" >"/tmp/v42-verify-$n.code") &
done
wait
CODES=$(cat /tmp/v42-verify-1.code /tmp/v42-verify-2.code | sort | tr '\n' ' ')
test "$CODES" = "202 409 "
test "$(sql "SELECT count(*) FROM business_verification_requests WHERE business_id='$BID' AND status='pending';")" = "1"

# Private verification material must not leak from account/admin responses.
ACCOUNT=$(curl -fsS -H 'X-User-Subject: v42-owner' "$API/v1/me/business-verifications")
ADMIN=$(curl -fsS -H "Authorization: Bearer $TOKEN" "$API/v1/admin/business-verifications?limit=100")
! grep -q 'private://v42-evidence' <<<"$ACCOUNT"
! grep -q 'private://v42-evidence' <<<"$ADMIN"
! grep -q '"evidence_ref"' <<<"$ACCOUNT"
! grep -q '"evidence_ref"' <<<"$ADMIN"

# Revoked ownership cannot create a verification request or mutate business state.
sql "UPDATE business_verification_requests SET status='cancelled' WHERE business_id='$BID' AND status='pending'; UPDATE businesses SET verification_status='unverified' WHERE id='$BID'; UPDATE business_ownerships SET status='revoked' WHERE business_id='$BID' AND user_id='$OWNER';"
BEFORE=$(sql "SELECT count(*) FROM business_verification_requests WHERE business_id='$BID';")
RC=$(code revoked /tmp/v42-revoked.json -H 'X-User-Subject: v42-owner' -H 'Content-Type: application/json' -d '{"method":"manual","statement":"must fail"}' "$API/v1/businesses/$BID/verification-requests")
test "$RC" = "403"
test "$(sql "SELECT count(*) FROM business_verification_requests WHERE business_id='$BID';")" = "$BEFORE"
test "$(sql "SELECT verification_status FROM businesses WHERE id='$BID';")" = "unverified"

# Rejected ownership is also denied.
sql "UPDATE business_ownerships SET status='rejected' WHERE business_id='$BID' AND user_id='$OWNER';"
RC=$(code rejected /tmp/v42-rejected.json -H 'X-User-Subject: v42-owner' -H 'Content-Type: application/json' -d '{"method":"manual","statement":"must fail"}' "$API/v1/businesses/$BID/verification-requests")
test "$RC" = "403"

# A stale pending request cannot be approved after ownership revocation; transaction must roll back.
sql "UPDATE business_ownerships SET status='pending' WHERE business_id='$BID' AND user_id='$OWNER'; UPDATE businesses SET verification_status='pending' WHERE id='$BID'; INSERT INTO business_verification_requests(business_id,user_id,method,statement,status) VALUES('$BID','$OWNER','manual','stale ownership','pending');"
VID=$(sql "SELECT id FROM business_verification_requests WHERE business_id='$BID' AND status='pending' ORDER BY created_at DESC LIMIT 1;")
sql "UPDATE business_ownerships SET status='revoked' WHERE business_id='$BID' AND user_id='$OWNER';"
RC=$(code stale-owner /tmp/v42-stale-owner.json -X PATCH -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"status":"approved"}' "$API/v1/admin/business-verifications/$VID")
test "$RC" = "409"
test "$(sql "SELECT status FROM business_verification_requests WHERE id='$VID';")" = "pending"
test "$(sql "SELECT verification_status FROM businesses WHERE id='$BID';")" = "pending"
sql "UPDATE business_verification_requests SET status='cancelled' WHERE id='$VID';"

# A stale request cannot be approved after business suspension.
sql "UPDATE business_ownerships SET status='pending' WHERE business_id='$BID' AND user_id='$OWNER'; UPDATE businesses SET verification_status='pending' WHERE id='$BID'; INSERT INTO business_verification_requests(business_id,user_id,method,statement,status) VALUES('$BID','$OWNER','manual','stale suspended','pending');"
VID=$(sql "SELECT id FROM business_verification_requests WHERE business_id='$BID' AND status='pending' ORDER BY created_at DESC LIMIT 1;")
sql "UPDATE businesses SET verification_status='suspended' WHERE id='$BID';"
RC=$(code stale-business /tmp/v42-stale-business.json -X PATCH -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"status":"approved"}' "$API/v1/admin/business-verifications/$VID")
test "$RC" = "409"
test "$(sql "SELECT status FROM business_verification_requests WHERE id='$VID';")" = "pending"
test "$(sql "SELECT verification_status FROM businesses WHERE id='$BID';")" = "suspended"

# Suspended business must be absent from public API and sitemap.
RC=$(curl -sS -o /tmp/v42-hidden.json -w '%{http_code}' "$API/v1/businesses/v42-ci-concurrent")
test "$RC" = "404"
! curl -fsS "$API/v1/seo/business-sitemap?limit=50000" | grep -q '"slug":"v42-ci-concurrent"'

# Concurrent reviews: exactly one active review per user/business.
sql "INSERT INTO businesses(slug,legal_name,display_name,verification_status,created_by) SELECT 'v42-ci-review','V42 Review','V42 Review','verified',id FROM user_accounts WHERE subject_key='v42-owner';"
RBID=$(sql "SELECT id FROM businesses WHERE slug='v42-ci-review';")
for n in 1 2; do
  (curl -sS -o "/tmp/v42-review-$n.json" -w '%{http_code}' -H 'X-User-Subject: v42-reviewer' -H 'Content-Type: application/json' -d '{"rating":5,"body":"V42 concurrent review"}' "$API/v1/businesses/$RBID/reviews" >"/tmp/v42-review-$n.code") &
done
wait
CODES=$(cat /tmp/v42-review-1.code /tmp/v42-review-2.code | sort | tr '\n' ' ')
test "$CODES" = "202 409 "
test "$(sql "SELECT count(*) FROM business_reviews WHERE business_id='$RBID' AND status IN ('pending','approved');")" = "1"
RID=$(sql "SELECT id FROM business_reviews WHERE business_id='$RBID' AND status='pending' LIMIT 1;")

# Moderation transition and replay.
RC=$(code review-ok /tmp/v42-review-ok.json -X PATCH -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"status":"approved"}' "$API/v1/admin/business-reviews/$RID")
test "$RC" = "200"
RC=$(code review-replay /tmp/v42-review-replay.json -X PATCH -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"status":"approved"}' "$API/v1/admin/business-reviews/$RID")
test "$RC" = "409"

# Verified public response contains approved review but no verification evidence fields.
PUBLIC=$(curl -fsS "$API/v1/businesses/v42-ci-review")
grep -q 'V42 concurrent review' <<<"$PUBLIC"
! grep -q '"evidence_ref"' <<<"$PUBLIC"
echo "V42 business hardening checks passed"
