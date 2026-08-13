package main

import (
	ent "github.com/troygilman/vent/examples/basic/ent"
	"github.com/troygilman/vent/examples/basic/ent/admin"
)

// AuthUserAdmin customizes the auth user admin surface.
// Embed DefaultAuthUserAdmin and override only what you need.
type AuthUserAdmin struct {
	admin.DefaultAuthUserAdmin
}

func (AuthUserAdmin) Name(e *ent.AuthUser) string {
	return e.Email
}

func (a AuthUserAdmin) FieldIsSuperuser() admin.AuthUserField {
	return SuperuserOnlyIsSuperuser{
		AuthUserIsSuperuserField: admin.NewAuthUserIsSuperuserField(a.Client),
	}
}
