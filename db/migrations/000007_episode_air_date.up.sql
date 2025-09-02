ALTER TABLE episodes RENAME COLUMN aired TO backup_aired;
ALTER TABLE episodes ADD COLUMN aired DATE;
UPDATE episodes SET aired = STR_TO_DATE(backup_aired, '%Y-%m-%d %H:%i:%s') WHERE backup_aired IS NOT NULL;
