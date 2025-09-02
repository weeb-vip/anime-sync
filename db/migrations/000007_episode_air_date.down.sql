/* drop the new aired (date) column */
ALTER TABLE episodes DROP COLUMN aired;

/* rename backup_aired back to aired */
ALTER TABLE episodes RENAME COLUMN backup_aired TO aired;
