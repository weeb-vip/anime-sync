CREATE TABLE anime_character_staff_link (
                                            id CHAR(36) NOT NULL PRIMARY KEY,
                                            character_id VARCHAR(36) NOT NULL,
                                            staff_id VARCHAR(36) NOT NULL,
                                            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
                                            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL
);
