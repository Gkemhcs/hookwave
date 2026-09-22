CREATE TABLE messages (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id  UUID         NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    event_type      VARCHAR(100) NOT NULL,
    payload         JSONB        NOT NULL,
    -- optional metadata for debugging (trace_id, source system etc.)
    metadata        JSONB,
    -- server-generated, consumer-facing dedup token (not producer-supplied,
    -- not unique-constrained — see idempotency-key discussion)
    idempotency_key VARCHAR(255) NOT NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);



