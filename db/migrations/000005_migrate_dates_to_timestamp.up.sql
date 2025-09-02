ALTER TABLE `anime` ADD `start_date2` TIMESTAMP NULL DEFAULT NULL AFTER `start_date`;
UPDATE `anime` SET start_date2=FROM_UNIXTIME(start_date) WHERE 1=1;
ALTER TABLE `anime` DROP `start_date`;
ALTER TABLE `anime` ADD `start_date` TIMESTAMP NULL DEFAULT NULL AFTER `start_date2`;
UPDATE `anime` SET start_date=start_date2 WHERE 1=1;
ALTER TABLE `anime` DROP `start_date2`;

ALTER TABLE `anime` ADD `end_date2` TIMESTAMP NULL DEFAULT NULL AFTER `end_date`;
UPDATE `anime` SET end_date2=FROM_UNIXTIME(end_date) WHERE 1=1;
ALTER TABLE `anime` DROP `end_date`;
ALTER TABLE `anime` ADD `end_date` TIMESTAMP NULL DEFAULT NULL AFTER `end_date2`;
UPDATE `anime` SET end_date=end_date2 WHERE 1=1;;
ALTER TABLE `anime` DROP `end_date2`;

ALTER TABLE `episodes` ADD `aired2` TIMESTAMP NULL DEFAULT NULL AFTER `aired`;
UPDATE `episodes` SET aired2=FROM_UNIXTIME(aired) WHERE 1=1;
ALTER TABLE `episodes` DROP `aired`;
ALTER TABLE `episodes` ADD `aired` TIMESTAMP NULL DEFAULT NULL AFTER `aired2`;
UPDATE `episodes` SET aired=aired2 WHERE 1=1;
ALTER TABLE `episodes` DROP `aired2`;



