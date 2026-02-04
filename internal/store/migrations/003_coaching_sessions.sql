-- +goose Up

CREATE TABLE coaching_sessions (
    id              BIGSERIAL PRIMARY KEY,
    puuid           VARCHAR(78) NOT NULL,
    latest_match_id VARCHAR(20) NOT NULL,
    match_ids       JSONB NOT NULL,
    analysis        JSONB NOT NULL,
    advice          TEXT NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_coaching_sessions_puuid_created ON coaching_sessions (puuid, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS coaching_sessions;
