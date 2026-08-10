-- Add column "last_login" to table: "auth_users"
ALTER TABLE `auth_users` ADD COLUMN `last_login` datetime NULL;
