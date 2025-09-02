ALTER TABLE anime_character_staff_link
    DROP INDEX idx_character_id,
    DROP INDEX idx_staff_id,
    DROP INDEX idx_character_staff;

ALTER TABLE anime_character
    DROP INDEX idx_anime_id;
