-- Create "permissions" table
CREATE TABLE `permissions` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `name` text NOT NULL);
-- Create index "permissions_name_key" to table: "permissions"
CREATE UNIQUE INDEX `permissions_name_key` ON `permissions` (`name`);
-- Create "permission_groups" table
CREATE TABLE `permission_groups` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `name` text NOT NULL);
-- Create index "permission_groups_name_key" to table: "permission_groups"
CREATE UNIQUE INDEX `permission_groups_name_key` ON `permission_groups` (`name`);
-- Create "users" table
CREATE TABLE `users` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `email` text NOT NULL, `password_hash` text NULL, `is_staff` bool NOT NULL DEFAULT (false), `is_superuser` bool NOT NULL DEFAULT (false), `is_active` bool NOT NULL DEFAULT (true), `last_login` datetime NULL);
-- Create index "users_email_key" to table: "users"
CREATE UNIQUE INDEX `users_email_key` ON `users` (`email`);
-- Create "permission_group_permissions" table
CREATE TABLE `permission_group_permissions` (`permission_group_id` integer NOT NULL, `permission_id` integer NOT NULL, PRIMARY KEY (`permission_group_id`, `permission_id`), CONSTRAINT `permission_group_permissions_permission_group_id` FOREIGN KEY (`permission_group_id`) REFERENCES `permission_groups` (`id`) ON DELETE CASCADE, CONSTRAINT `permission_group_permissions_permission_id` FOREIGN KEY (`permission_id`) REFERENCES `permissions` (`id`) ON DELETE CASCADE);
-- Create "user_groups" table
CREATE TABLE `user_groups` (`user_id` integer NOT NULL, `permission_group_id` integer NOT NULL, PRIMARY KEY (`user_id`, `permission_group_id`), CONSTRAINT `user_groups_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE, CONSTRAINT `user_groups_permission_group_id` FOREIGN KEY (`permission_group_id`) REFERENCES `permission_groups` (`id`) ON DELETE CASCADE);
