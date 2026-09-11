-- Create "authors" table
CREATE TABLE `authors` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `active` bool NOT NULL DEFAULT (true), `user_id` integer NOT NULL, CONSTRAINT `authors_users_author` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE NO ACTION);
-- Create index "authors_user_id_key" to table: "authors"
CREATE UNIQUE INDEX `authors_user_id_key` ON `authors` (`user_id`);
-- Create "books" table
CREATE TABLE `books` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `title` text NOT NULL, `pages` integer NOT NULL DEFAULT (0), `published` bool NOT NULL DEFAULT (false), `published_at` datetime NULL, `created_at` datetime NOT NULL, `internal_notes` text NULL, `book_author` integer NOT NULL, CONSTRAINT `books_authors_author` FOREIGN KEY (`book_author`) REFERENCES `authors` (`id`) ON DELETE NO ACTION);
-- Create "permissions" table
CREATE TABLE `permissions` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `name` text NOT NULL);
-- Create index "permissions_name_key" to table: "permissions"
CREATE UNIQUE INDEX `permissions_name_key` ON `permissions` (`name`);
-- Create "permission_groups" table
CREATE TABLE `permission_groups` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `name` text NOT NULL);
-- Create index "permission_groups_name_key" to table: "permission_groups"
CREATE UNIQUE INDEX `permission_groups_name_key` ON `permission_groups` (`name`);
-- Create "reviews" table
CREATE TABLE `reviews` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `rating` integer NOT NULL, `body` text NULL, `book_reviews` integer NOT NULL, `user_reviews` integer NOT NULL, CONSTRAINT `reviews_books_reviews` FOREIGN KEY (`book_reviews`) REFERENCES `books` (`id`) ON DELETE NO ACTION, CONSTRAINT `reviews_users_reviews` FOREIGN KEY (`user_reviews`) REFERENCES `users` (`id`) ON DELETE NO ACTION);
-- Create "users" table
CREATE TABLE `users` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `username` text NOT NULL, `password_hash` text NULL, `is_staff` bool NOT NULL DEFAULT (false), `is_superuser` bool NOT NULL DEFAULT (false), `is_active` bool NOT NULL DEFAULT (true), `last_login` datetime NULL);
-- Create index "users_username_key" to table: "users"
CREATE UNIQUE INDEX `users_username_key` ON `users` (`username`);
-- Create "permission_group_permissions" table
CREATE TABLE `permission_group_permissions` (`permission_group_id` integer NOT NULL, `permission_id` integer NOT NULL, PRIMARY KEY (`permission_group_id`, `permission_id`), CONSTRAINT `permission_group_permissions_permission_group_id` FOREIGN KEY (`permission_group_id`) REFERENCES `permission_groups` (`id`) ON DELETE CASCADE, CONSTRAINT `permission_group_permissions_permission_id` FOREIGN KEY (`permission_id`) REFERENCES `permissions` (`id`) ON DELETE CASCADE);
-- Create "user_permission_groups" table
CREATE TABLE `user_permission_groups` (`user_id` integer NOT NULL, `permission_group_id` integer NOT NULL, PRIMARY KEY (`user_id`, `permission_group_id`), CONSTRAINT `user_permission_groups_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE, CONSTRAINT `user_permission_groups_permission_group_id` FOREIGN KEY (`permission_group_id`) REFERENCES `permission_groups` (`id`) ON DELETE CASCADE);
-- Seed generated auth permissions
INSERT INTO `permissions` (`name`) VALUES ('read_author'), ('create_author'), ('update_author'), ('delete_author'), ('read_book'), ('create_book'), ('update_book'), ('delete_book'), ('publish'), ('read_permission'), ('update_permission'), ('read_permission_group'), ('create_permission_group'), ('update_permission_group'), ('delete_permission_group'), ('read_review'), ('create_review'), ('update_review'), ('read_user'), ('create_user'), ('update_user'), ('delete_user'), ('impersonate');
