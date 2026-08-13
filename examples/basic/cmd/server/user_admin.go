package main

import (
	ent "github.com/troygilman/vent/examples/basic/ent"
	"github.com/troygilman/vent/examples/basic/ent/admin"
)

// UserAdmin customizes the user admin surface.
// Embed DefaultUserAdmin and override only what you need.
type UserAdmin struct {
	admin.DefaultUserAdmin
}

func (UserAdmin) Name(e *ent.User) string {
	return e.Email
}

func (a UserAdmin) FieldIsSuperuser() admin.UserField {
	return SuperuserOnlyIsSuperuser{
		UserIsSuperuserField: admin.NewUserIsSuperuserField(a.Client),
	}
}
