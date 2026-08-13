-- Removed permissions
DELETE FROM `permissions` WHERE `name` IN ('create_permission', 'delete_permission');
