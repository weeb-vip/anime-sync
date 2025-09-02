ALTER TABLE anime_character_staff_link
    ADD COLUMN character_name VARCHAR(255) NOT NULL AFTER staff_id,
    ADD COLUMN staff_given_name VARCHAR(255) NOT NULL AFTER character_name,
    ADD COLUMN staff_family_name VARCHAR(255) NOT NULL AFTER staff_given_name;
