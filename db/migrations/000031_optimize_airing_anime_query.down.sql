-- Rollback airing anime query optimization

DROP INDEX IF EXISTS idx_episodes_aired_anime_id ON episodes;
DROP INDEX IF EXISTS idx_anime_end_date_id ON anime;