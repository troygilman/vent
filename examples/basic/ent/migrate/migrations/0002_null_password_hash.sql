-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_auth_users" table
CREATE TABLE `new_auth_users` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `email` text NOT NULL, `password_hash` text NULL, `is_staff` bool NOT NULL DEFAULT (false), `is_superuser` bool NOT NULL DEFAULT (false), `is_active` bool NOT NULL DEFAULT (true));
-- Copy rows from old table "auth_users" to new temporary table "new_auth_users"
INSERT INTO `new_auth_users` (`id`, `email`, `password_hash`, `is_staff`, `is_superuser`, `is_active`) SELECT `id`, `email`, `password_hash`, `is_staff`, `is_superuser`, `is_active` FROM `auth_users`;
-- Drop "auth_users" table after copying rows
DROP TABLE `auth_users`;
-- Rename temporary table "new_auth_users" to "auth_users"
ALTER TABLE `new_auth_users` RENAME TO `auth_users`;
-- Create index "auth_users_email_key" to table: "auth_users"
CREATE UNIQUE INDEX `auth_users_email_key` ON `auth_users` (`email`);
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
