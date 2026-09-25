CREATE TABLE businesses (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 slug VARCHAR(180) NOT NULL UNIQUE,
 legal_name VARCHAR(200) NOT NULL,
 display_name VARCHAR(200) NOT NULL,
 description TEXT,
 website_url TEXT,
 verification_status VARCHAR(24) NOT NULL DEFAULT 'unverified' CHECK(verification_status IN ('unverified','pending','verified','rejected','suspended')),
 created_by UUID REFERENCES user_accounts(id) ON DELETE SET NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_businesses_verified ON businesses(updated_at DESC) WHERE verification_status='verified';

CREATE TABLE business_branches (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 business_id UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
 name VARCHAR(200) NOT NULL,
 address_text TEXT,
 region_code VARCHAR(64),
 latitude NUMERIC(9,6),
 longitude NUMERIC(9,6),
 is_primary BOOLEAN NOT NULL DEFAULT FALSE,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_business_branches_business ON business_branches(business_id,is_primary DESC,created_at);

CREATE TABLE business_phone_links (
 business_id UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
 phone_number_id UUID NOT NULL REFERENCES phone_numbers(id) ON DELETE CASCADE,
 branch_id UUID REFERENCES business_branches(id) ON DELETE CASCADE,
 relationship VARCHAR(32) NOT NULL DEFAULT 'business' CHECK(relationship IN ('business','hotline','support','sales','branch')),
 is_public BOOLEAN NOT NULL DEFAULT TRUE,
 verified_at TIMESTAMPTZ,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 PRIMARY KEY(business_id,phone_number_id)
);
CREATE INDEX idx_business_phone_links_phone ON business_phone_links(phone_number_id) WHERE is_public=TRUE;

CREATE TABLE business_ownerships (
 business_id UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
 user_id UUID NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
 role VARCHAR(24) NOT NULL DEFAULT 'owner' CHECK(role IN ('owner','manager')),
 status VARCHAR(24) NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','verified','rejected','revoked')),
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 verified_at TIMESTAMPTZ,
 PRIMARY KEY(business_id,user_id)
);

CREATE TABLE business_verification_requests (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 business_id UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
 user_id UUID NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
 method VARCHAR(32) NOT NULL CHECK(method IN ('phone','email','document','manual')),
 statement TEXT,
 evidence_ref TEXT,
 status VARCHAR(24) NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','approved','rejected','cancelled')),
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 reviewed_at TIMESTAMPTZ,
 reviewed_by VARCHAR(128),
 resolution_note TEXT
);
CREATE INDEX idx_business_verification_queue ON business_verification_requests(status,created_at DESC);
CREATE UNIQUE INDEX idx_business_verification_one_pending ON business_verification_requests(business_id,user_id) WHERE status='pending';
