-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_users" table
CREATE TABLE `new_users` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `username` text NOT NULL, `password_hash` text NULL, `is_staff` bool NOT NULL DEFAULT (false), `is_superuser` bool NOT NULL DEFAULT (false), `is_active` bool NOT NULL DEFAULT (true), `last_login` datetime NULL);
-- Copy rows from old table "users" to new temporary table "new_users"
INSERT INTO `new_users` (`id`, `username`, `password_hash`, `is_staff`, `is_superuser`, `is_active`, `last_login`) SELECT `id`, `email`, `password_hash`, `is_staff`, `is_superuser`, `is_active`, `last_login` FROM `users`;
-- Drop "users" table after copying rows
DROP TABLE `users`;
-- Rename temporary table "new_users" to "users"
ALTER TABLE `new_users` RENAME TO `users`;
-- Create index "users_username_key" to table: "users"
CREATE UNIQUE INDEX `users_username_key` ON `users` (`username`);
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
