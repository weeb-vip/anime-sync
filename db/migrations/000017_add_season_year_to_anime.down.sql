ALTER TABLE anime DROP INDEX idx_anime_season_year;
ALTER TABLE anime DROP INDEX idx_anime_year;
ALTER TABLE anime DROP INDEX idx_anime_season;

ALTER TABLE anime DROP COLUMN year;
ALTER TABLE anime DROP COLUMN season;