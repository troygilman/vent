package main

import (
	"context"

	"github.com/troygilman/vent"
	"github.com/troygilman/vent/examples/basic/ent"
	"github.com/troygilman/vent/examples/basic/ent/admin"
	"github.com/troygilman/vent/examples/basic/ent/user"
	"github.com/troygilman/vent/templates/gui"
)

// SuperuserOnlyIsSuperuser restricts is_superuser changes to superusers.
// It embeds the generated default so list behavior stays unchanged.
type SuperuserOnlyIsSuperuser struct {
	admin.UserIsSuperuserField
}

func (f SuperuserOnlyIsSuperuser) CreateHTML(ctx context.Context) (string, error) {
	editable, err := currentUserIsSuperuser(ctx)
	if err != nil {
		return "", err
	}
	return gui.RenderBoolFieldHTML(ctx, gui.SchemaEntityBoolFieldProps{
		Name:     "is_superuser",
		Label:    "IsSuperuser",
		Value:    vent.FormatFormValue(user.DefaultIsSuperuser),
		Editable: editable && gui.MustRenderContext(ctx).CanUpdate,
	})
}

func (f SuperuserOnlyIsSuperuser) UpdateHTML(ctx context.Context, e *ent.User) (string, error) {
	editable, err := currentUserIsSuperuser(ctx)
	if err != nil {
		return "", err
	}
	return gui.RenderBoolFieldHTML(ctx, gui.SchemaEntityBoolFieldProps{
		Name:     "is_superuser",
		Label:    "IsSuperuser",
		Value:    vent.FormatFormValue(e.IsSuperuser),
		Editable: editable && gui.MustRenderContext(ctx).CanUpdate,
	})
}

func (f SuperuserOnlyIsSuperuser) ApplyCreate(ctx context.Context, builder *ent.UserCreate, input admin.UserCreateInput) error {
	if err := requireSuperuserForSuperuserFlag(ctx, input.IsSuperuser != nil); err != nil {
		return err
	}
	return f.UserIsSuperuserField.ApplyCreate(ctx, builder, input)
}

func (f SuperuserOnlyIsSuperuser) ApplyUpdate(ctx context.Context, builder *ent.UserUpdateOne, input admin.UserUpdateInput) error {
	if err := requireSuperuserForSuperuserFlag(ctx, input.IsSuperuser != nil); err != nil {
		return err
	}
	return f.UserIsSuperuserField.ApplyUpdate(ctx, builder, input)
}

func currentUserIsSuperuser(ctx context.Context) (bool, error) {
	user, err := admin.GetUser(ctx)
	if err != nil {
		return false, err
	}
	return user.IsSuperuser, nil
}

func requireSuperuserForSuperuserFlag(ctx context.Context, changing bool) error {
	if !changing {
		return nil
	}
	ok, err := currentUserIsSuperuser(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return vent.Forbidden("only superusers can change is_superuser")
	}
	return nil
}
