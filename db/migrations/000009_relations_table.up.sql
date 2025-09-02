CREATE TABLE anime_relations (
                                 id CHAR(36) NOT NULL PRIMARY KEY,                  -- store UUID as string
                                 anime_id VARCHAR(36) NOT NULL,
                                 related_anime_id VARCHAR(36) NOT NULL,
                                 relation_type VARCHAR(30),                         -- e.g., 'sequel', 'prequel', etc.
                                 created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
