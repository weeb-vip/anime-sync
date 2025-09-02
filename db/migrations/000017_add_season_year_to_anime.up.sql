ALTER TABLE anime ADD COLUMN season VARCHAR(20) NULL;
ALTER TABLE anime ADD COLUMN year INTEGER NULL;

CREATE INDEX idx_anime_season ON anime(season);
CREATE INDEX idx_anime_year ON anime(year);
CREATE INDEX idx_anime_season_year ON anime(season, year);