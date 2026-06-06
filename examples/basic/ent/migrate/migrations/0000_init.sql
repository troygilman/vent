-- Create "auth_groups" table
CREATE TABLE `auth_groups` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `name` text NOT NULL);
-- Create index "auth_groups_name_key" to table: "auth_groups"
CREATE UNIQUE INDEX `auth_groups_name_key` ON `auth_groups` (`name`);
-- Create "auth_permissions" table
CREATE TABLE `auth_permissions` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `name` text NOT NULL);
-- Create index "auth_permissions_name_key" to table: "auth_permissions"
CREATE UNIQUE INDEX `auth_permissions_name_key` ON `auth_permissions` (`name`);
-- Create "auth_users" table
CREATE TABLE `auth_users` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `email` text NOT NULL, `password_hash` text NOT NULL, `is_staff` bool NOT NULL DEFAULT (false), `is_superuser` bool NOT NULL DEFAULT (false), `is_active` bool NOT NULL DEFAULT (true));
-- Create index "auth_users_email_key" to table: "auth_users"
CREATE UNIQUE INDEX `auth_users_email_key` ON `auth_users` (`email`);
-- Create "auth_group_permissions" table
CREATE TABLE `auth_group_permissions` (`auth_group_id` integer NOT NULL, `auth_permission_id` integer NOT NULL, PRIMARY KEY (`auth_group_id`, `auth_permission_id`), CONSTRAINT `auth_group_permissions_auth_group_id` FOREIGN KEY (`auth_group_id`) REFERENCES `auth_groups` (`id`) ON DELETE CASCADE, CONSTRAINT `auth_group_permissions_auth_permission_id` FOREIGN KEY (`auth_permission_id`) REFERENCES `auth_permissions` (`id`) ON DELETE CASCADE);
-- Create "auth_user_groups" table
CREATE TABLE `auth_user_groups` (`auth_user_id` integer NOT NULL, `auth_group_id` integer NOT NULL, PRIMARY KEY (`auth_user_id`, `auth_group_id`), CONSTRAINT `auth_user_groups_auth_user_id` FOREIGN KEY (`auth_user_id`) REFERENCES `auth_users` (`id`) ON DELETE CASCADE, CONSTRAINT `auth_user_groups_auth_group_id` FOREIGN KEY (`auth_group_id`) REFERENCES `auth_groups` (`id`) ON DELETE CASCADE);
