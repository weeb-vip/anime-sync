ALTER TABLE anime_character_staff_link
    ADD CONSTRAINT fk_character_id FOREIGN KEY (character_id) REFERENCES anime_character(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_staff_id FOREIGN KEY (staff_id) REFERENCES anime_staff(id) ON DELETE CASCADE;
