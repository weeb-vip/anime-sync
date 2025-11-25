-- Remove database indexes

-- Drop procedure if it exists first, then create it
DROP PROCEDURE IF EXISTS DropIndexIfExists;

-- Helper procedure to safely drop indexes
DELIMITER //
CREATE PROCEDURE DropIndexIfExists(
    IN indexName VARCHAR(255),
    IN tableName VARCHAR(255)
)
BEGIN
    DECLARE index_exists INT DEFAULT 0;
    
    SELECT COUNT(*) INTO index_exists
    FROM information_schema.statistics 
    WHERE table_schema = DATABASE() 
    AND table_name = tableName 
    AND index_name = indexName;
    
    IF index_exists > 0 THEN
        SET @sql = CONCAT('DROP INDEX ', indexName, ' ON ', tableName);
        PREPARE stmt FROM @sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END IF;
END//
DELIMITER ;

-- Anime table indexes
CALL DropIndexIfExists('idx_anime_status', 'anime');
CALL DropIndexIfExists('idx_anime_rating', 'anime');
CALL DropIndexIfExists('idx_anime_ranking', 'anime');
CALL DropIndexIfExists('idx_anime_created_at', 'anime');
CALL DropIndexIfExists('idx_anime_type', 'anime');
CALL DropIndexIfExists('idx_anime_source', 'anime');
CALL DropIndexIfExists('idx_anime_start_date', 'anime');
CALL DropIndexIfExists('idx_anime_end_date', 'anime');
CALL DropIndexIfExists('idx_anime_anidbid', 'anime');
CALL DropIndexIfExists('idx_anime_thetvdbid', 'anime');

-- Full-text search indexes for title fields
CALL DropIndexIfExists('idx_anime_title_en', 'anime');
CALL DropIndexIfExists('idx_anime_title_jp', 'anime');
CALL DropIndexIfExists('idx_anime_title_romaji', 'anime');
CALL DropIndexIfExists('idx_anime_title_kanji', 'anime');

-- Episodes table indexes
CALL DropIndexIfExists('idx_episodes_anime_id', 'episodes');
CALL DropIndexIfExists('idx_episodes_aired', 'episodes');
CALL DropIndexIfExists('idx_episodes_episode_number', 'episodes');
CALL DropIndexIfExists('idx_episodes_created_at', 'episodes');

-- Composite index for episodes by anime and air date
CALL DropIndexIfExists('idx_episodes_anime_aired', 'episodes');

-- Anime characters table indexes
CALL DropIndexIfExists('idx_anime_character_anime_id', 'anime_character');
CALL DropIndexIfExists('idx_anime_character_name', 'anime_character');
CALL DropIndexIfExists('idx_anime_character_role', 'anime_character');

-- Anime staff table indexes
CALL DropIndexIfExists('idx_anime_staff_given_name', 'anime_staff');
CALL DropIndexIfExists('idx_anime_staff_family_name', 'anime_staff');
CALL DropIndexIfExists('idx_anime_staff_language', 'anime_staff');

-- Anime character staff link table indexes
CALL DropIndexIfExists('idx_character_staff_character_id', 'anime_character_staff_link');
CALL DropIndexIfExists('idx_character_staff_staff_id', 'anime_character_staff_link');

-- Anime seasons table indexes
CALL DropIndexIfExists('idx_anime_seasons_episode_count', 'anime_seasons');
CALL DropIndexIfExists('idx_anime_seasons_created_at', 'anime_seasons');

-- Relations table indexes
CALL DropIndexIfExists('idx_relations_anime_id', 'relations');
CALL DropIndexIfExists('idx_relations_related_anime_id', 'relations');

-- Drop the helper procedure
DROP PROCEDURE DropIndexIfExists;