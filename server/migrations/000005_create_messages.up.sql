CREATE TABLE messages (
    id                UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id    UUID         NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    event_type        VARCHAR(100) NOT NULL,
    payload           JSONB        NOT NULL,
    metadata          JSONB,
    idempotency_key   VARCHAR(255),
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(application_id, idempotency_key)
);
