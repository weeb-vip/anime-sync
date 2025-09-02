CREATE TABLE anime_character (
                                 id CHAR(36) NOT NULL PRIMARY KEY,
                                 anime_id VARCHAR(36) NOT NULL,
                                 name VARCHAR(255) NOT NULL,
                                 role VARCHAR(255) NOT NULL,
                                 birthday VARCHAR(255),
                                 zodiac VARCHAR(255),
                                 gender VARCHAR(255),
                                 race VARCHAR(255),
                                 height VARCHAR(255),
                                 weight VARCHAR(255),
                                 title VARCHAR(255),
                                 martial_status VARCHAR(255),
                                 summary TEXT,
                                 image TEXT,
                                 created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
                                 updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL
);
