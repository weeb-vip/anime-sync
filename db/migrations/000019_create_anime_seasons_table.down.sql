-- Add back old season column to anime table
ALTER TABLE anime ADD COLUMN season VARCHAR(50) NULL;

-- Drop indexes
ALTER TABLE anime_seasons DROP INDEX IDX_anime_id;
ALTER TABLE anime_seasons DROP INDEX IDX_status;
ALTER TABLE anime_seasons DROP INDEX IDX_season;

-- Drop anime_seasons table
DROP TABLE IF EXISTS anime_seasons;