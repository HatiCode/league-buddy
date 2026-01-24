-- +goose Up
-- League Buddy initial schema

CREATE TABLE summoners (
    id BIGSERIAL PRIMARY KEY,
    puuid VARCHAR(78) NOT NULL UNIQUE,
    summoner_id VARCHAR(63) NOT NULL,
    name VARCHAR(16) NOT NULL,
    platform VARCHAR(10) NOT NULL,
    profile_icon_id INTEGER DEFAULT 0,
    summoner_level BIGINT DEFAULT 0,
    tier VARCHAR(12) DEFAULT '',
    rank VARCHAR(4) DEFAULT '',
    league_points INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_summoners_platform_name ON summoners (platform, LOWER(name));

CREATE TABLE matches (
    id BIGSERIAL PRIMARY KEY,
    match_id VARCHAR(20) NOT NULL UNIQUE,
    platform VARCHAR(10) NOT NULL,
    queue_id INTEGER NOT NULL,
    game_mode VARCHAR(20) NOT NULL,
    game_duration BIGINT NOT NULL,
    game_version VARCHAR(20) NOT NULL,
    game_ended_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE participants (
    id BIGSERIAL PRIMARY KEY,
    match_id BIGINT NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    puuid VARCHAR(78) NOT NULL,
    summoner_name VARCHAR(16) NOT NULL,
    champion_id INTEGER NOT NULL,
    champion_name VARCHAR(30) NOT NULL,
    team_id INTEGER NOT NULL,
    team_position VARCHAR(10) NOT NULL,
    win BOOLEAN NOT NULL,
    kills INTEGER NOT NULL DEFAULT 0,
    deaths INTEGER NOT NULL DEFAULT 0,
    assists INTEGER NOT NULL DEFAULT 0,
    total_minions_killed INTEGER NOT NULL DEFAULT 0,
    neutral_minions_killed INTEGER NOT NULL DEFAULT 0,
    vision_score INTEGER NOT NULL DEFAULT 0,
    wards_placed INTEGER NOT NULL DEFAULT 0,
    wards_killed INTEGER NOT NULL DEFAULT 0,
    detector_wards_placed INTEGER NOT NULL DEFAULT 0,
    damage_dealt INTEGER NOT NULL DEFAULT 0,
    damage_taken INTEGER NOT NULL DEFAULT 0,
    gold_earned INTEGER NOT NULL DEFAULT 0,
    dragon_kills INTEGER NOT NULL DEFAULT 0,
    baron_kills INTEGER NOT NULL DEFAULT 0,
    turret_kills INTEGER NOT NULL DEFAULT 0,
    first_blood_kill BOOLEAN NOT NULL DEFAULT FALSE,
    first_blood_assist BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE(match_id, puuid)
);

CREATE INDEX idx_participants_puuid ON participants (puuid);

CREATE TABLE summoner_matches (
    summoner_id BIGINT NOT NULL REFERENCES summoners(id) ON DELETE CASCADE,
    match_id BIGINT NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (summoner_id, match_id)
);

CREATE INDEX idx_summoner_matches_summoner_created ON summoner_matches (summoner_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS summoner_matches;
DROP TABLE IF EXISTS participants;
DROP TABLE IF EXISTS matches;
DROP TABLE IF EXISTS summoners;
