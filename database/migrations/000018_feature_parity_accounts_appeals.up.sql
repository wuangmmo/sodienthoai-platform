CREATE TABLE user_phone_preferences (
 user_id UUID NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
 phone_number_id UUID NOT NULL REFERENCES phone_numbers(id) ON DELETE CASCADE,
 disposition VARCHAR(16) NOT NULL CHECK(disposition IN ('trusted','blocked')),
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 PRIMARY KEY(user_id,phone_number_id)
);
CREATE INDEX idx_user_phone_preferences_user_disposition
 ON user_phone_preferences(user_id,disposition,updated_at DESC);

CREATE TABLE user_lookup_history (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 user_id UUID NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
 phone_number_id UUID REFERENCES phone_numbers(id) ON DELETE SET NULL,
 e164 VARCHAR(32) NOT NULL,
 looked_up_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_user_lookup_history_user_time
 ON user_lookup_history(user_id,looked_up_at DESC);
CREATE INDEX idx_user_lookup_history_phone
 ON user_lookup_history(phone_number_id)
 WHERE phone_number_id IS NOT NULL;

CREATE TABLE phone_appeals (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 phone_number_id UUID NOT NULL REFERENCES phone_numbers(id) ON DELETE CASCADE,
 user_id UUID NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
 reason VARCHAR(64) NOT NULL,
 statement TEXT NOT NULL CHECK(char_length(statement) BETWEEN 10 AND 4000),
 status VARCHAR(24) NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','accepted','rejected','withdrawn')),
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 reviewed_at TIMESTAMPTZ,
 reviewed_by VARCHAR(128),
 resolution_note TEXT
);
CREATE INDEX idx_phone_appeals_moderation
 ON phone_appeals(status,created_at DESC);
CREATE UNIQUE INDEX idx_phone_appeals_one_pending
 ON phone_appeals(phone_number_id,user_id)
 WHERE status='pending';
