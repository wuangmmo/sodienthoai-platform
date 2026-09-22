CREATE TABLE user_accounts (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 subject_key VARCHAR(128) NOT NULL UNIQUE,
 display_name VARCHAR(120),
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE phone_follows (
 user_id UUID NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
 phone_number_id UUID NOT NULL REFERENCES phone_numbers(id) ON DELETE CASCADE,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 PRIMARY KEY(user_id,phone_number_id)
);
CREATE INDEX idx_phone_follows_phone ON phone_follows(phone_number_id);

CREATE TABLE phone_comments (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 phone_number_id UUID NOT NULL REFERENCES phone_numbers(id) ON DELETE CASCADE,
 user_id UUID NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
 body TEXT NOT NULL CHECK(char_length(body) BETWEEN 1 AND 1000),
 status VARCHAR(24) NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','approved','rejected','removed')),
 helpful_count INTEGER NOT NULL DEFAULT 0,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 reviewed_at TIMESTAMPTZ
);
CREATE INDEX idx_phone_comments_public ON phone_comments(phone_number_id,created_at DESC) WHERE status='approved';

CREATE TABLE user_notifications (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 user_id UUID NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
 phone_number_id UUID REFERENCES phone_numbers(id) ON DELETE CASCADE,
 event_type VARCHAR(64) NOT NULL,
 payload JSONB NOT NULL DEFAULT '{}'::jsonb,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 read_at TIMESTAMPTZ
);
CREATE INDEX idx_user_notifications_unread ON user_notifications(user_id,created_at DESC) WHERE read_at IS NULL;
