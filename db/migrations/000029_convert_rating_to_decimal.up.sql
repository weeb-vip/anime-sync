-- Convert rating column from TEXT to DECIMAL for better performance

-- Step 1: Add temporary column for conversion
ALTER TABLE anime ADD COLUMN rating_temp DECIMAL(3,1) DEFAULT NULL;

-- Step 2: Populate the temporary column with converted data
UPDATE anime
SET rating_temp = CASE
    -- Handle empty or N/A ratings
    WHEN rating IS NULL OR rating = '' OR rating = 'N/A' OR rating = 'Unknown' THEN NULL

    -- Handle numeric ratings (most common case: "8.5", "9.2", etc.)
    WHEN rating REGEXP '^[0-9]+\\.?[0-9]*$' THEN
        CASE
            WHEN CAST(rating AS DECIMAL(3,1)) BETWEEN 0.0 AND 10.0 THEN CAST(rating AS DECIMAL(3,1))
            ELSE NULL
        END

    -- Handle ratings with extra text (e.g., "8.5/10", "9.2 stars")
    WHEN rating REGEXP '^([0-9]+\\.?[0-9]*)' THEN
        CASE
            WHEN CAST(SUBSTRING_INDEX(rating, '/', 1) AS DECIMAL(3,1)) BETWEEN 0.0 AND 10.0
            THEN CAST(SUBSTRING_INDEX(rating, '/', 1) AS DECIMAL(3,1))
            ELSE NULL
        END

    -- Default case for invalid formats
    ELSE NULL
END;

-- Step 3: Drop the old rating column
ALTER TABLE anime DROP COLUMN rating;

-- Step 4: Rename the temporary column to rating
ALTER TABLE anime CHANGE COLUMN rating_temp rating DECIMAL(3,1) DEFAULT NULL;

-- Step 5: Create optimized index for rating queries
CREATE INDEX idx_anime_rating_desc ON anime (rating DESC);

-- Step 6: Create composite index for stable sorting
CREATE INDEX idx_anime_rating_id ON anime (rating DESC, id);