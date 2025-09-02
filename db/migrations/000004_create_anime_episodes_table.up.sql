CREATE TABLE episodes
(
    id         varchar(36) PRIMARY KEY,
    anime_id   varchar(36) REFERENCES anime (id),
    episode    int,
    title_en   text,
    title_jp   text,
    aired      timestamp,
    synopsis   text,
    created_at timestamp,
    updated_at timestamp
);
