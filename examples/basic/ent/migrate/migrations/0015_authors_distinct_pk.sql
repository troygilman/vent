-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_authors" table
CREATE TABLE `new_authors` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `active` bool NOT NULL DEFAULT (true), `user_id` integer NOT NULL, CONSTRAINT `authors_users_author` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE NO ACTION);
-- Copy rows: former shared PK value becomes the user_id FK
INSERT INTO `new_authors` (`id`, `active`, `user_id`) SELECT `id`, `active`, `id` FROM `authors`;
-- Drop "authors" table after copying rows
DROP TABLE `authors`;
-- Rename temporary table "new_authors" to "authors"
ALTER TABLE `new_authors` RENAME TO `authors`;
-- Create index "authors_user_id_key" to table: "authors"
CREATE UNIQUE INDEX `authors_user_id_key` ON `authors` (`user_id`);
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
