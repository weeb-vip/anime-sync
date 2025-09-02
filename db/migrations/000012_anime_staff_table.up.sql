CREATE TABLE anime_staff (
                             id CHAR(36) NOT NULL PRIMARY KEY,
                             given_name VARCHAR(255) NOT NULL,
                             family_name VARCHAR(255) NOT NULL,
                             image TEXT,
                             birthday VARCHAR(255),
                             birth_place VARCHAR(255),
                             blood_type VARCHAR(255),
                             hobbies VARCHAR(255),
                             summary TEXT,
                             created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
                             updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL
);
