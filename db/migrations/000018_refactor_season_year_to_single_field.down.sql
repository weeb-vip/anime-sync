-- Drop the new combined column and its index
ALTER TABLE anime DROP INDEX idx_anime_season;
ALTER TABLE anime DROP COLUMN season;

-- Re-add the separate season and year columns
ALTER TABLE anime ADD COLUMN season VARCHAR(20) NULL;
ALTER TABLE anime ADD COLUMN year INTEGER NULL;

-- Re-create the old indexes
CREATE INDEX idx_anime_season ON anime(season);
CREATE INDEX idx_anime_year ON anime(year);
CREATE INDEX idx_anime_season_year ON anime(season, year);