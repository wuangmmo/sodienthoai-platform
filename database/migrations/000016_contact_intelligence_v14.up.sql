CREATE TABLE contact_books (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 owner_key VARCHAR(128) NOT NULL,
 name VARCHAR(160) NOT NULL DEFAULT 'Danh bạ của tôi',
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_contact_books_owner ON contact_books(owner_key,created_at DESC);

CREATE TABLE private_contacts (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 contact_book_id UUID NOT NULL REFERENCES contact_books(id) ON DELETE CASCADE,
 local_name VARCHAR(255),
 raw_phone VARCHAR(80) NOT NULL,
 normalized_e164 VARCHAR(32),
 phone_number_id UUID REFERENCES phone_numbers(id) ON DELETE SET NULL,
 match_status identity_match_status NOT NULL DEFAULT 'unknown',
 local_note TEXT,
 source VARCHAR(32) NOT NULL DEFAULT 'upload' CHECK(source IN ('upload','manual','sync')),
 last_checked_at TIMESTAMPTZ,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 UNIQUE(contact_book_id,normalized_e164)
);
CREATE INDEX idx_private_contacts_book_status ON private_contacts(contact_book_id,match_status);
CREATE INDEX idx_private_contacts_phone ON private_contacts(phone_number_id);

CREATE TABLE contact_identity_actions (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 contact_id UUID NOT NULL REFERENCES private_contacts(id) ON DELETE CASCADE,
 action_type VARCHAR(40) NOT NULL CHECK(action_type IN ('keep_local','update_local','suggest_public_correction','report_inaccurate','invite_owner')),
 payload JSONB NOT NULL DEFAULT '{}'::jsonb,
 status VARCHAR(24) NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','completed','cancelled')),
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 completed_at TIMESTAMPTZ
);
CREATE INDEX idx_contact_identity_actions_contact ON contact_identity_actions(contact_id,created_at DESC);
