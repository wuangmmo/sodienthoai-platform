CREATE TABLE business_reviews (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 business_id UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
 branch_id UUID REFERENCES business_branches(id) ON DELETE SET NULL,
 user_id UUID NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
 rating SMALLINT NOT NULL CHECK(rating BETWEEN 1 AND 5),
 body TEXT CHECK(body IS NULL OR char_length(body) BETWEEN 1 AND 2000),
 status VARCHAR(24) NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','approved','rejected','removed')),
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 reviewed_at TIMESTAMPTZ,
 reviewed_by VARCHAR(128)
);
CREATE INDEX idx_business_reviews_public ON business_reviews(business_id,created_at DESC) WHERE status='approved';
CREATE INDEX idx_business_reviews_moderation ON business_reviews(status,created_at DESC);
CREATE UNIQUE INDEX idx_business_reviews_user_business_active ON business_reviews(business_id,user_id) WHERE status IN ('pending','approved');

CREATE TABLE business_review_reports (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 review_id UUID NOT NULL REFERENCES business_reviews(id) ON DELETE CASCADE,
 user_id UUID NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
 reason VARCHAR(64) NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 UNIQUE(review_id,user_id)
);
