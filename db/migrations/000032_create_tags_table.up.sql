-- Create tags table
CREATE TABLE tags
(
    id         BIGINT AUTO_INCREMENT PRIMARY KEY,
    name       VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_tags_name (name)
);

-- Create anime_tags junction table for many-to-many relationship
CREATE TABLE anime_tags
(
    anime_id   VARCHAR(36) NOT NULL,
    tag_id     BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (anime_id, tag_id),
    FOREIGN KEY (anime_id) REFERENCES anime(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE,
    INDEX idx_anime_tags_anime_id (anime_id),
    INDEX idx_anime_tags_tag_id (tag_id)
);

-- Migrate existing genres data to tags table
-- Genres are stored as JSON arrays like: ["Drama","Supernatural","Suspense"]
INSERT INTO tags (name)
SELECT DISTINCT TRIM(REPLACE(REPLACE(j.genre, '"', ''), '[', '')) AS tag_name
FROM anime a
CROSS JOIN JSON_TABLE(
    a.genres,
    '$[*]' COLUMNS (genre VARCHAR(100) PATH '$')
) AS j
WHERE a.genres IS NOT NULL
  AND JSON_VALID(a.genres)
  AND TRIM(j.genre) != ''
ON DUPLICATE KEY UPDATE name = name;

-- Populate anime_tags junction table from existing genres data
INSERT INTO anime_tags (anime_id, tag_id)
SELECT DISTINCT a.id, t.id
FROM anime a
CROSS JOIN JSON_TABLE(
    a.genres,
    '$[*]' COLUMNS (genre VARCHAR(100) PATH '$')
) AS j
INNER JOIN tags t ON t.name = TRIM(j.genre)
WHERE a.genres IS NOT NULL
  AND JSON_VALID(a.genres)
  AND TRIM(j.genre) != '';
