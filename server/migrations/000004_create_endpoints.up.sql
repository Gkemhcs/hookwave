CREATE TABLE endpoints (
    id              UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100)  NOT NULL,
    application_id  UUID          NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    url             VARCHAR(2048) NOT NULL,
    method          VARCHAR(10)   NOT NULL DEFAULT 'POST',
    headers         JSONB,
    query_params    JSONB,
    event_types     JSONB         NOT NULL DEFAULT '[]',
    signing_secret  VARCHAR(255)  NOT NULL,
    is_active       BOOLEAN       NOT NULL DEFAULT true,
    timeout_ms      INT           NOT NULL DEFAULT 5000,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);
