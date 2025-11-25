-- Add missing composite indexes for optimal anime ranking query performance

-- Composite index for MostPopularAnime: ORDER BY ranking asc, id
-- Lower ranking = more popular (ranking 1 is better than ranking 100)
CREATE INDEX idx_anime_ranking_id ON anime (ranking ASC, id);

-- Composite index for NewestAnime: ORDER BY created_at desc, id
-- Newer anime first (most recent created_at first)
CREATE INDEX idx_anime_created_at_id ON anime (created_at DESC, id);

-- Note: Rating composite index (idx_anime_rating_id) is already created in migration 000029