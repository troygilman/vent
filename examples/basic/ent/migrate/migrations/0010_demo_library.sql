-- Create "api_keys" table
CREATE TABLE `api_keys` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `name` text NOT NULL, `token` text NOT NULL);
-- Create index "api_keys_name_key" to table: "api_keys"
CREATE UNIQUE INDEX `api_keys_name_key` ON `api_keys` (`name`);
-- Create "audit_events" table
CREATE TABLE `audit_events` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `action` text NOT NULL, `detail` text NULL, `created_at` datetime NOT NULL);
-- Create "authors" table
CREATE TABLE `authors` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `name` text NOT NULL, `bio` text NULL, `active` bool NOT NULL DEFAULT (true));
-- Create "books" table
CREATE TABLE `books` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `title` text NOT NULL, `isbn` text NULL, `pages` integer NOT NULL DEFAULT (0), `price` real NOT NULL DEFAULT (0), `published` bool NOT NULL DEFAULT (false), `published_at` datetime NULL, `created_at` datetime NOT NULL, `view_count` integer NOT NULL DEFAULT (0), `internal_notes` text NULL, `book_author` integer NOT NULL, `book_category` integer NULL, CONSTRAINT `books_authors_author` FOREIGN KEY (`book_author`) REFERENCES `authors` (`id`) ON DELETE NO ACTION, CONSTRAINT `books_categories_category` FOREIGN KEY (`book_category`) REFERENCES `categories` (`id`) ON DELETE SET NULL);
-- Create "categories" table
CREATE TABLE `categories` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `name` text NOT NULL, `description` text NULL);
-- Create index "categories_name_key" to table: "categories"
CREATE UNIQUE INDEX `categories_name_key` ON `categories` (`name`);
-- Create "reviews" table
CREATE TABLE `reviews` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `reviewer` text NOT NULL, `rating` integer NOT NULL, `body` text NULL, `created_at` datetime NOT NULL, `book_reviews` integer NOT NULL, CONSTRAINT `reviews_books_reviews` FOREIGN KEY (`book_reviews`) REFERENCES `books` (`id`) ON DELETE NO ACTION);
-- Create "tags" table
CREATE TABLE `tags` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `name` text NOT NULL);
-- Create index "tags_name_key" to table: "tags"
CREATE UNIQUE INDEX `tags_name_key` ON `tags` (`name`);
-- Create "book_tags" table
CREATE TABLE `book_tags` (`book_id` integer NOT NULL, `tag_id` integer NOT NULL, PRIMARY KEY (`book_id`, `tag_id`), CONSTRAINT `book_tags_book_id` FOREIGN KEY (`book_id`) REFERENCES `books` (`id`) ON DELETE CASCADE, CONSTRAINT `book_tags_tag_id` FOREIGN KEY (`tag_id`) REFERENCES `tags` (`id`) ON DELETE CASCADE);
