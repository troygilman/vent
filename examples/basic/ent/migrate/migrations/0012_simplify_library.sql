-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Drop unused example tables
DROP TABLE IF EXISTS `book_tags`;
DROP TABLE IF EXISTS `tags`;
DROP TABLE IF EXISTS `api_keys`;
DROP TABLE IF EXISTS `audit_events`;
DROP TABLE IF EXISTS `categories`;
-- Create "new_authors" table
CREATE TABLE `new_authors` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `name` text NOT NULL, `active` bool NOT NULL DEFAULT (true));
-- Copy rows from old table "authors" to new temporary table "new_authors"
INSERT INTO `new_authors` (`id`, `name`, `active`) SELECT `id`, `name`, `active` FROM `authors`;
-- Drop "authors" table after copying rows
DROP TABLE `authors`;
-- Rename temporary table "new_authors" to "authors"
ALTER TABLE `new_authors` RENAME TO `authors`;
-- Create "new_books" table
CREATE TABLE `new_books` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `title` text NOT NULL, `pages` integer NOT NULL DEFAULT (0), `published` bool NOT NULL DEFAULT (false), `published_at` datetime NULL, `created_at` datetime NOT NULL, `internal_notes` text NULL, `book_author` integer NOT NULL, CONSTRAINT `books_authors_author` FOREIGN KEY (`book_author`) REFERENCES `authors` (`id`) ON DELETE NO ACTION);
-- Copy rows from old table "books" to new temporary table "new_books"
INSERT INTO `new_books` (`id`, `title`, `pages`, `published`, `published_at`, `created_at`, `internal_notes`, `book_author`) SELECT `id`, `title`, `pages`, `published`, `published_at`, `created_at`, `internal_notes`, `book_author` FROM `books`;
-- Drop "books" table after copying rows
DROP TABLE `books`;
-- Rename temporary table "new_books" to "books"
ALTER TABLE `new_books` RENAME TO `books`;
-- Create "new_reviews" table
CREATE TABLE `new_reviews` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `reviewer` text NOT NULL, `rating` integer NOT NULL, `body` text NULL, `book_reviews` integer NOT NULL, CONSTRAINT `reviews_books_reviews` FOREIGN KEY (`book_reviews`) REFERENCES `books` (`id`) ON DELETE NO ACTION);
-- Copy rows from old table "reviews" to new temporary table "new_reviews"
INSERT INTO `new_reviews` (`id`, `reviewer`, `rating`, `body`, `book_reviews`) SELECT `id`, `reviewer`, `rating`, `body`, `book_reviews` FROM `reviews`;
-- Drop "reviews" table after copying rows
DROP TABLE `reviews`;
-- Rename temporary table "new_reviews" to "reviews"
ALTER TABLE `new_reviews` RENAME TO `reviews`;
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
