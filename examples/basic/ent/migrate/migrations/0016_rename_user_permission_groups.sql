-- Drop "user_groups" table (old User.groups join table; membership rows are not copied)
DROP TABLE IF EXISTS `user_groups`;
-- Create "user_permission_groups" table
CREATE TABLE `user_permission_groups` (`user_id` integer NOT NULL, `permission_group_id` integer NOT NULL, PRIMARY KEY (`user_id`, `permission_group_id`), CONSTRAINT `user_permission_groups_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE, CONSTRAINT `user_permission_groups_permission_group_id` FOREIGN KEY (`permission_group_id`) REFERENCES `permission_groups` (`id`) ON DELETE CASCADE);
