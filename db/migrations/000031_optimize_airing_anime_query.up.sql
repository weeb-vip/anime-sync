-- Optimize CurrentlyAiring query performance

-- Create composite index to handle the complex WHERE condition more efficiently
-- This index supports: anime.end_date IS NULL OR anime.end_date >= ?
CREATE INDEX idx_anime_end_date_id ON anime (end_date, id);

-- Create covering index for episodes that includes both WHERE and SELECT needs
-- This allows the episodes subquery to be resolved entirely from the index
CREATE INDEX idx_episodes_aired_anime_id ON episodes (aired, anime_id);