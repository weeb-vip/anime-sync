-- Rollback composite indexes

-- Drop the composite indexes
DROP INDEX IF EXISTS idx_anime_created_at_id ON anime;
DROP INDEX IF EXISTS idx_anime_ranking_id ON anime;