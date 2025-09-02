ALTER TABLE anime_character_staff_link
    ADD INDEX idx_character_id (character_id),
    ADD INDEX idx_staff_id (staff_id),
    ADD INDEX idx_character_staff (character_id, staff_id);

ALTER TABLE anime_character
    ADD INDEX idx_anime_id (anime_id);
