/* rename aired to backup_aired */
ALTER TABLE episodes RENAME COLUMN aired TO backup_aired;
/* add aired column with type date */
ALTER TABLE episodes ADD COLUMN aired date;

/* update aired column with the value from backup_aired */
UPDATE episodes SET aired = to_timestamp(backup_aired, 'YYYY-MM-DD HH24:MI:SS')::date;