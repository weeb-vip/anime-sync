CREATE TABLE episode_air_time (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    anime_id VARCHAR(36) NOT NULL,
    episode_number INT NOT NULL,
    air_type ENUM('raw', 'sub', 'dub') NOT NULL,
    air_datetime TIMESTAMP NOT NULL,
    streams_json JSON NULL,
    last_synced_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_episode_air_time_air_datetime (air_datetime),
    UNIQUE INDEX idx_episode_air_time_unique (anime_id, episode_number, air_type),
    FOREIGN KEY (anime_id) REFERENCES anime(id) ON DELETE CASCADE
);
