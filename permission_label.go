package vent

// FormatPermissionSelectLabel builds the admin selector label for a permission.
// The stored permission name is preserved and prefixed with its schema.
func FormatPermissionSelectLabel(schemaDisplay, permissionName string) string {
	if schemaDisplay == "" {
		return permissionName
	}
	return schemaDisplay + " | " + permissionName
}
