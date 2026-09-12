CREATE TYPE delivery_status AS ENUM ('pending', 'delivered', 'failed', 'dead_lettered');

CREATE TABLE deliveries (
    id                UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id        UUID            NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    endpoint_id       UUID            NOT NULL REFERENCES endpoints(id) ON DELETE CASCADE,
    status            delivery_status NOT NULL DEFAULT 'pending',
    attempts          INT             NOT NULL DEFAULT 0,
    max_attempts      INT             NOT NULL DEFAULT 5,
    next_attempt_at   TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    locked_until      TIMESTAMPTZ,
    last_attempted_at TIMESTAMPTZ,
    -- last HTTP response from endpoint, useful for debugging
    response_status   INT,
    response_body     TEXT,
    created_at        TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- partial index for SKIP LOCKED claim query
-- only indexes pending rows — stays small as delivered/dead_lettered rows accumulate
CREATE INDEX idx_deliveries_claim
    ON deliveries (next_attempt_at, locked_until)
    WHERE status = 'pending';
