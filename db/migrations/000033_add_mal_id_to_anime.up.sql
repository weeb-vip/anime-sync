ALTER TABLE anime ADD COLUMN mal_id INT NULL AFTER anidbid;
CREATE INDEX idx_anime_mal_id ON anime(mal_id);
