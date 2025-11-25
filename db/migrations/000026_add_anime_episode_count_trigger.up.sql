CREATE TRIGGER update_anime_episode_count_after_insert
AFTER INSERT ON episodes
FOR EACH ROW
BEGIN
    UPDATE anime
    SET episodes = (
        SELECT COUNT(*)
        FROM episodes
        WHERE anime_id = NEW.anime_id
    )
    WHERE id = NEW.anime_id;
END;