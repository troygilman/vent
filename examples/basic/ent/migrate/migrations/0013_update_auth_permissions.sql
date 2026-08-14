-- Remove permissions for dropped example schemas
DELETE FROM `permissions` WHERE `name` IN ('read_audit_event', 'read_category', 'create_category', 'update_category', 'delete_category', 'read_tag', 'create_tag', 'update_tag', 'delete_tag');
