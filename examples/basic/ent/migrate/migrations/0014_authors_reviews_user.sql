-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Existing demo rows cannot map Author.id onto users.id or Review onto a user.
DROP TABLE IF EXISTS `reviews`;
DROP TABLE IF EXISTS `books`;
DROP TABLE IF EXISTS `authors`;
-- Create "authors" table (id is both PK and FK to users)
CREATE TABLE `authors` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `active` bool NOT NULL DEFAULT (true), CONSTRAINT `authors_users_author` FOREIGN KEY (`id`) REFERENCES `users` (`id`) ON DELETE NO ACTION);
-- Create "books" table
CREATE TABLE `books` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `title` text NOT NULL, `pages` integer NOT NULL DEFAULT (0), `published` bool NOT NULL DEFAULT (false), `published_at` datetime NULL, `created_at` datetime NOT NULL, `internal_notes` text NULL, `book_author` integer NOT NULL, CONSTRAINT `books_authors_author` FOREIGN KEY (`book_author`) REFERENCES `authors` (`id`) ON DELETE NO ACTION);
-- Create "reviews" table
CREATE TABLE `reviews` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `rating` integer NOT NULL, `body` text NULL, `book_reviews` integer NOT NULL, `user_reviews` integer NOT NULL, CONSTRAINT `reviews_books_reviews` FOREIGN KEY (`book_reviews`) REFERENCES `books` (`id`) ON DELETE NO ACTION, CONSTRAINT `reviews_users_reviews` FOREIGN KEY (`user_reviews`) REFERENCES `users` (`id`) ON DELETE NO ACTION);
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
