CREATE TRIGGER update_anime_episode_count_after_delete
AFTER DELETE ON episodes
FOR EACH ROW
BEGIN
    UPDATE anime
    SET episodes = (
        SELECT COUNT(*)
        FROM episodes
        WHERE anime_id = OLD.anime_id
    )
    WHERE id = OLD.anime_id;
END;