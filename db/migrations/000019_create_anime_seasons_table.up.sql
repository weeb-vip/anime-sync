-- Create anime_seasons table
CREATE TABLE anime_seasons (
    id            VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    season        VARCHAR(255) NOT NULL,
    status        ENUM('unknown', 'confirmed', 'announced', 'cancelled') DEFAULT 'unknown' NOT NULL,
    episode_count INTEGER,
    notes         TEXT,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL,
    anime_id      VARCHAR(36),
    CONSTRAINT UQ_anime_season UNIQUE (anime_id, season)
);

-- Create indexes
CREATE INDEX IDX_season ON anime_seasons (season);
CREATE INDEX IDX_status ON anime_seasons (status);
CREATE INDEX IDX_anime_id ON anime_seasons (anime_id);

-- Remove old season column from anime table
ALTER TABLE anime DROP COLUMN season;