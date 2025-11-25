-- Add database indexes for improved query performance
-- Using CREATE INDEX directly (indexes will fail silently if they already exist)

-- Anime table indexes (TEXT columns need key length specification)
CREATE INDEX idx_anime_status ON anime (status(50));
CREATE INDEX idx_anime_ranking ON anime (ranking);
CREATE INDEX idx_anime_created_at ON anime (created_at);
CREATE INDEX idx_anime_type ON anime (type);
CREATE INDEX idx_anime_source ON anime (source(100));
CREATE INDEX idx_anime_start_date ON anime (start_date);
CREATE INDEX idx_anime_end_date ON anime (end_date);
CREATE INDEX idx_anime_anidbid ON anime (anidbid);
CREATE INDEX idx_anime_thetvdbid ON anime (thetvdbid);

-- Full-text search indexes for title fields
CREATE INDEX idx_anime_title_en ON anime (title_en(255));
CREATE INDEX idx_anime_title_jp ON anime (title_jp(255));
CREATE INDEX idx_anime_title_romaji ON anime (title_romaji(255));
CREATE INDEX idx_anime_title_kanji ON anime (title_kanji(255));

-- Episodes table indexes
CREATE INDEX idx_episodes_anime_id ON episodes (anime_id);
CREATE INDEX idx_episodes_aired ON episodes (aired);
CREATE INDEX idx_episodes_episode_number ON episodes (episode);
CREATE INDEX idx_episodes_created_at ON episodes (created_at);

-- Composite index for episodes by anime and air date
CREATE INDEX idx_episodes_anime_aired ON episodes (anime_id, aired);

-- Anime characters table indexes
CREATE INDEX idx_anime_character_anime_id ON anime_character (anime_id);
CREATE INDEX idx_anime_character_name ON anime_character (name);
CREATE INDEX idx_anime_character_role ON anime_character (role);

-- Anime staff table indexes
CREATE INDEX idx_anime_staff_given_name ON anime_staff (given_name);
CREATE INDEX idx_anime_staff_family_name ON anime_staff (family_name);
CREATE INDEX idx_anime_staff_language ON anime_staff (language);

-- Anime character staff link table indexes
CREATE INDEX idx_character_staff_character_id ON anime_character_staff_link (character_id);
CREATE INDEX idx_character_staff_staff_id ON anime_character_staff_link (staff_id);

-- Anime seasons table indexes
CREATE INDEX idx_anime_seasons_episode_count ON anime_seasons (episode_count);
CREATE INDEX idx_anime_seasons_created_at ON anime_seasons (created_at);

-- Anime relations table indexes
CREATE INDEX idx_anime_relations_anime_id ON anime_relations (anime_id);
CREATE INDEX idx_anime_relations_related_anime_id ON anime_relations (related_anime_id);
