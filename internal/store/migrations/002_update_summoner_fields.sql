-- +goose Up
-- Remove deprecated fields and add Riot ID fields

ALTER TABLE summoners DROP COLUMN summoner_id;
ALTER TABLE summoners DROP COLUMN name;
DROP INDEX IF EXISTS idx_summoners_platform_name;

ALTER TABLE summoners ADD COLUMN game_name VARCHAR(24) NOT NULL DEFAULT '';
ALTER TABLE summoners ADD COLUMN tag_line VARCHAR(8) NOT NULL DEFAULT '';
ALTER TABLE summoners ADD COLUMN revision_date BIGINT NOT NULL DEFAULT 0;

CREATE INDEX idx_summoners_platform_riot_id ON summoners (platform, LOWER(game_name), LOWER(tag_line));

-- +goose Down
DROP INDEX IF EXISTS idx_summoners_platform_riot_id;

ALTER TABLE summoners DROP COLUMN game_name;
ALTER TABLE summoners DROP COLUMN tag_line;
ALTER TABLE summoners DROP COLUMN revision_date;

ALTER TABLE summoners ADD COLUMN summoner_id VARCHAR(63) NOT NULL DEFAULT '';
ALTER TABLE summoners ADD COLUMN name VARCHAR(16) NOT NULL DEFAULT '';

CREATE INDEX idx_summoners_platform_name ON summoners (platform, LOWER(name));
