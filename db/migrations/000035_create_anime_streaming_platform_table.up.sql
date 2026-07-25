CREATE TABLE anime_streaming_platform (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    anime_id VARCHAR(36) NOT NULL,
    platform VARCHAR(100) NOT NULL,
    name VARCHAR(255) NULL,
    url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE INDEX idx_streaming_anime_platform (anime_id, platform),
    FOREIGN KEY (anime_id) REFERENCES anime(id) ON DELETE CASCADE
);
