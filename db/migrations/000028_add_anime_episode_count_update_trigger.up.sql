CREATE TRIGGER update_anime_episode_count_after_update
AFTER UPDATE ON episodes
FOR EACH ROW
BEGIN
    IF OLD.anime_id != NEW.anime_id THEN
        -- Update old anime's count
        UPDATE anime
        SET episodes = (
            SELECT COUNT(*)
            FROM episodes
            WHERE anime_id = OLD.anime_id
        )
        WHERE id = OLD.anime_id;

        -- Update new anime's count
        UPDATE anime
        SET episodes = (
            SELECT COUNT(*)
            FROM episodes
            WHERE anime_id = NEW.anime_id
        )
        WHERE id = NEW.anime_id;
    END IF;
END;