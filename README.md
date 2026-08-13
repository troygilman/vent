# Vent

A lightweight, type-safe admin panel and CRUD framework for Go. Vent generates a fully functional admin interface from your [Ent](https://entgo.io/) schemas with almost zero boilerplate.

## Features

- **Zero-config admin UI** — register an Ent extension and get a complete admin panel
- **Automatic CRUD** for any Ent schema
- **Built-in auth** — users, groups, and permissions with RBAC (JWT + bcrypt)
- **Customizable** via schema annotations
- **Lightweight & embeddable** — great for internal tools

## Tech Stack

| Layer    | Technology                                |
| -------- | ----------------------------------------- |
| Backend  | Go + Ent                                  |
| UI       | Templ + [Datastar](https://datastar.dev/) |
| Styling  | Native CSS (custom properties)            |
| Database | Any Ent-supported DB (SQLite demo)        |

## Quick Start

1. **Install**

    ```bash
    go get github.com/troygilman/vent
    ```

2. **Scaffold the auth schemas** using the `vent` CLI:

    ```bash
    go run ./cmd/vent init --schema ./ent/schema
    ```

    This writes `user.go`, `permission_group.go`, and `permission.go` into your schema directory. Each schema uses the corresponding Vent mixin:

    ```go
    func (User) Mixin() []ent.Mixin {
        return []ent.Mixin{
            vent.UserMixin{GroupSchemaType: PermissionGroup.Type},
        }
    }
    ```

3. **Generate Ent + admin code:**

    ```bash
    go run ./cmd/vent gen --schema ./ent/schema
    ```

4. **Mount the admin handler** in your server:

    ```go
    adminHandler, err := admin.NewAdminHandler(admin.AdminConfig{
        Client:                 client,
        SecretProvider:         auth.SecretProviderFunc(func() []byte { return []byte("secret") }),
        CredentialGenerator:    auth.NewBCryptCredentialGenerator(),
        CredentialAuthenticator: auth.NewBCryptCredentialAuthenticator(),
    })
    mux := http.NewServeMux()
    mux.Handle("/admin/", adminHandler)
    ```

    Then visit `http://localhost:8080/admin/`.

## Customizing schemas

Generated admin code exposes `AdminConfig.Schemas` (`SchemaAdmins`), with one `*Admin` interface slot per schema. Nil means “use the generated `Default*Admin`.”

Each schema admin owns:

- **`FieldX()`** — one method per surface member (e.g. `FieldEmail()`, `FieldIsSuperuser()`)
- **`ValidateCreate` / `ValidateUpdate` / `ValidateDelete`** — mutation policy
- **`CanRead` / `CanCreate` / `CanUpdate` / `CanDelete`** — permission checks for routes, nav, and UI controls

Embed the default and override only what you need:

```go
type UserAdmin struct {
    admin.DefaultUserAdmin
}

func (a UserAdmin) FieldIsSuperuser() admin.UserField {
    return SuperuserOnlyIsSuperuser{
        UserIsSuperuserField: admin.NewUserIsSuperuserField(a.Client),
    }
}

// other Field* + Validate* inherited from DefaultUserAdmin

adminHandler, err := admin.NewAdminHandler(admin.AdminConfig{
    Client: client,
    // ...auth deps...
    Schemas: admin.SchemaAdmins{
        User: UserAdmin{
            DefaultUserAdmin: admin.NewDefaultUserAdmin(client),
        },
    },
})
```

To layer validation on the default policy:

```go
func (a UserAdmin) ValidateUpdate(ctx context.Context, id int, in admin.UserUpdateInput) error {
    if err := a.DefaultUserAdmin.ValidateUpdate(ctx, id, in); err != nil {
        return err
    }
    // additional checks...
    return nil
}
```

Declare virtual custom fields on the schema with `VentSchemaAnnotation.CustomFields`, then implement the matching `FieldX()` method (required when there is no generated default).

Keep app types **outside** `ent/admin` — that package is regenerated and non-keep files are removed.

See [`examples/basic/cmd/server/user_admin.go`](examples/basic/cmd/server/user_admin.go) and [`superuser_field.go`](examples/basic/cmd/server/superuser_field.go).

## Example

A complete working example with SQLite lives in [`examples/basic/`](examples/basic/). Run it with:

```bash
just dev   # generate + run the example server
just gen   # regenerate Templ, CSS, and Vent admin code
```

---

**Vent** — _Breathe life into your admin panels._
