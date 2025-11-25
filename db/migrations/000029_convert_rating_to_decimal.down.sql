-- Rollback rating column conversion from DECIMAL to TEXT

-- Step 1: Drop the indexes
DROP INDEX IF EXISTS idx_anime_rating_id ON anime;
DROP INDEX IF EXISTS idx_anime_rating_desc ON anime;

-- Step 2: Add temporary TEXT column
ALTER TABLE anime ADD COLUMN rating_temp TEXT DEFAULT NULL;

-- Step 3: Convert DECIMAL back to TEXT (with some data loss for non-standard formats)
UPDATE anime
SET rating_temp = CASE
    WHEN rating IS NULL THEN 'N/A'
    ELSE CAST(rating AS CHAR)
END;

-- Step 4: Drop the DECIMAL rating column
ALTER TABLE anime DROP COLUMN rating;

-- Step 5: Rename temporary column back to rating
ALTER TABLE anime CHANGE COLUMN rating_temp rating TEXT DEFAULT NULL;