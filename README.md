# Vent

A lightweight, type-safe admin panel and CRUD framework for Go. Vent turns your [Ent](https://entgo.io/) schemas into a fully functional admin UI with almost zero boilerplate — codegen for handlers and fields, built-in auth and RBAC, and a Templ + Datastar front end you can embed in any Go HTTP server.

## Features

- **Zero-config admin UI** — register the Vent Ent extension and get list, create, edit, and delete screens for every schema
- **Schema annotations** — control routes, labels, table columns, fieldsets, read-only mode, and custom permissions without hand-writing handlers
- **Per-schema customization** — override fields, validation, and `Can*` checks by embedding generated defaults
- **Built-in auth** — users, permission groups, and permissions with JWT sessions, bcrypt credentials, CSRF, and staff/superuser gates
- **Custom fields** — declare virtual form fields (including the built-in `password` field on auth users) and implement them like any other field
- **Permission migrator** — keep the permission table in sync with generated CRUD + custom permission names
- **Lightweight & embeddable** — static assets and UI templates ship with the module; mount under any path

## Tech stack

| Layer    | Technology                                |
| -------- | ----------------------------------------- |
| Backend  | Go + Ent                                  |
| UI       | Templ + [Datastar](https://datastar.dev/) |
| Styling  | Native CSS (custom properties)            |
| Database | Any Ent-supported DB (SQLite in the demo) |

---

## How it works

Vent extends Ent’s codegen. You describe your models as Ent schemas (plus Vent mixins and annotations); Vent generates an `ent/admin` package with HTTP handlers, form fields, and permission metadata. You mount that handler in your server and get a working admin UI.

```text
Ent schemas  →  vent gen  →  ent/admin  →  mount under /admin/
```

Customization stays in your app code: annotate schemas for layout and labels, or embed the generated `Default*Admin` types to override fields, validation, and permissions. Auth is built in — staff users sign in with JWT sessions, and access is controlled through permission groups (with superuser bypass).

---

## Setup

### 1. Install

```bash
go get github.com/troygilman/vent
```

You need a working Ent project (`ent/schema`, codegen target, etc.).

### 2. Scaffold auth schemas

```bash
go run github.com/troygilman/vent/cmd/vent init --schema ./ent/schema
```

Writes `user.go`, `permission_group.go`, and `permission.go`. Each schema uses the matching Vent mixin:

```go
func (User) Mixin() []ent.Mixin {
    return []ent.Mixin{
        vent.UserMixin{GroupSchemaType: PermissionGroup.Type},
    }
}
```

Use `--force` only when you intentionally want to overwrite those files.

You may add extra fields/edges on those schemas (see the example `User.last_login`). Keep the mixins and auth roles intact — codegen validates that the configured auth schemas use the correct `VentAuthMixin` roles.

### 3. Generate Ent + admin code

```bash
go run github.com/troygilman/vent/cmd/vent gen --schema ./ent/schema
```

Options:

| Flag | Default | Purpose |
| ---- | ------- | ------- |
| `--schema` / `-s` | `./ent/schema` | Ent schema directory |
| `--admin-path` | `/admin/` | Mount path baked into generated routes |

`gen` runs Ent with Vent’s admin extension and enables versioned migrations, upsert, and snapshot features. Generated admin sources land in `ent/admin/`. That package is regenerated on every run — **do not edit it**; put overrides in your application package.

To customize auth schema type names or call the extension from your own `entc` program:

```go
entc.Generate("./ent/schema", &gen.Config{ /* ... */ },
    entc.Extensions(vent.NewAdminExtension(
        vent.WithAdminPath("/admin/"),
        vent.WithAuthSchemas(vent.AuthSchemas{
            User:       schema.User.Type,
            Group:      schema.PermissionGroup.Type,
            Permission: schema.Permission.Type,
        }),
    )),
)
```

(`vent gen` currently exposes `--admin-path` only; use a custom `entc` main for `WithAuthSchemas`.)

### 4. Migrate the database

Apply your Ent/Atlas migrations as usual. When schemas or custom permissions change, regenerate migrations so permission rows stay current. The example project does:

```bash
just migrations   # NamedDiff + admin.Diff for permission rows
just migrate      # atlas migrate apply ...
```

`admin.Diff` compares the live permission set to the generated list and writes an `update_auth_permissions` migration when needed. Today the permission differ is SQLite-oriented; use the same dialect as your Ent migrations for schema changes.

### 5. Mount the admin handler

```go
adminHandler, err := admin.NewAdminHandler(admin.AdminConfig{
    Client:                  client,
    SecretProvider:          auth.SecretProviderFunc(func() []byte { return []byte(os.Getenv("VENT_JWT_SECRET")) }),
    CredentialGenerator:     auth.NewBCryptCredentialGenerator(),
    CredentialAuthenticator: auth.NewBCryptCredentialAuthenticator(),
    SecureCookies:           true, // set in production (HTTPS)
    // Schemas: optional overrides; nil slots use Default*Admin
})
if err != nil {
    log.Fatal(err)
}

mux := http.NewServeMux()
mux.Handle("/admin/", adminHandler)
```

Create at least one staff/superuser so you can sign in. The example seeds `admin@vent.com` / `test_user`.

Then open `http://localhost:8080/admin/`.

### Production checklist

- Use a strong, rotatable JWT secret via `SecretProvider` (read on every generate/authenticate — do not snapshot at construction)
- Enable `SecureCookies` behind HTTPS
- Prefer bcrypt (or your own `Credential*` implementations) for password hashing
- Treat client-facing errors as public messages only; use `vent.HttpError` / `vent.HandleError` so internal causes stay in logs

---

## Auth and permissions

Vent expects three auth schemas (defaults: `User`, `PermissionGroup`, `Permission`) that use the corresponding mixins. Access is layered:

| Gate | Behavior |
| ---- | -------- |
| Session | Valid JWT; user must be `is_active` |
| Staff | `is_staff` required for the admin UI |
| Schema CRUD | Permissions named `read_<resource>`, `create_<resource>`, `update_<resource>`, `delete_<resource>` |
| Superuser | `is_superuser` bypasses permission checks; only superusers may mutate other superusers |
| Entity `Can*` | Optional per-row policy on top of schema permissions |

Permissions are granted through **permission groups**. Custom names declared in `VentSchemaAnnotation.Permissions` are inserted by the permission migrator (they are data you can assign to groups; wire them into route middleware yourself if you need enforcement beyond the default CRUD set).

Built-in safety on the auth user schema:

- Users cannot delete or deactivate themselves
- Clearing your own password is rejected
- Non-superusers cannot update or delete superusers

Mutating requests are CSRF-protected (cookie + `X-CSRF-Token` header). Theme preference is stored separately (`system` / `light` / `dark`).

---

## Schema annotations

Annotate any Ent schema with `vent.VentSchemaAnnotation` (mixins already attach sensible defaults for auth schemas). Schema-level annotations replace mixin defaults entirely today (deep-merge is planned).

```go
func (Book) Annotations() []schema.Annotation {
    return []schema.Annotation{
        vent.VentSchemaAnnotation{
            RouteName:           "books",
            SingularDisplayName: "Book",
            PluralDisplayName:   "Books",
            TableColumns:        []string{"title", "author", "published"},
            FilterableColumns:   []string{"title", "published"},
            FieldSets: []vent.FieldSet{{
                Fields: []string{"title", "author", "published", "tags"},
            }},
            ReadOnlyFields: []string{"published"},
            Permissions: []vent.Permission{
                {Name: "publish", Desc: "Publish a book"},
            },
            CustomFields: []vent.Field{
                {Name: "notes", Type: "string", InputType: "string"},
            },
        },
    }
}
```

| Field | Effect |
| ----- | ------ |
| `DisableAdmin` | Skip admin UI and routes for this schema |
| `ReadOnly` | Disable create, update, and delete |
| `DisableCreate` / `DisableDelete` | Turn off individual mutation routes |
| `ReadOnlyFields` | Show on forms but do not bind on create/update |
| `RouteName` | URL segment (default: pluralized resource name) |
| `SingularDisplayName` / `PluralDisplayName` | Nav and page titles |
| `TableColumns` | List-view columns (fields or edges) |
| `FilterableColumns` | List-view filters for string, bool, and int fields |
| `FieldSets` | Form field order (first set is used; multi-set UI is incomplete) |
| `CustomFields` | Virtual surface members you implement via `FieldX()` |
| `Permissions` | Extra permission rows (name + description) for the migrator |

Supported form/input kinds: `string`, `password`, `int` (and width variants), `float`, `bool`, `time`, `foreign_key`, `foreign_key_unique`. Edges render as FK selectors (unique vs multi).

---

## Customizing the admin surface

Generated code exposes `AdminConfig.Schemas` (`SchemaAdmins`) with one `*Admin` interface per schema. A **nil** slot means “use `Default*Admin`.”

Each `*Admin` owns:

- **`FieldX()`** — one method per surface member
- **`Name(entity)`** — display label (lists, breadcrumbs, current-user chip)
- **`ValidateCreate` / `ValidateUpdate` / `ValidateDelete`** — mutation policy after bind, before save
- **`CanRead` / `CanCreate` / `CanUpdate` / `CanDelete`** — permission checks for routes, nav, and UI controls

Keep app types **outside** `ent/admin`. Embed the default and override only what you need:

```go
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

Layer validation on the default policy:

```go
func (a UserAdmin) ValidateUpdate(ctx context.Context, id int, in admin.UserUpdateInput) error {
    if err := a.DefaultUserAdmin.ValidateUpdate(ctx, id, in); err != nil {
        return err
    }
    // additional checks...
    return nil
}
```

### Field contract

Every field implements the same interface (name varies by schema), for example:

```go
type UserField interface {
    ListCell(ctx context.Context, e *ent.User) string
    CreateHTML(ctx context.Context) (string, error)
    UpdateHTML(ctx context.Context, e *ent.User) (string, error)
    ApplyCreate(ctx context.Context, builder *ent.UserCreate, input UserCreateInput) error
    ApplyUpdate(ctx context.Context, builder *ent.UserUpdateOne, input UserUpdateInput) error
}
```

Generated defaults cover Ent fields, edges, and the built-in `password` custom field. User-declared `CustomFields` without a built-in default **must** supply `FieldX()` — `NewAdminHandler` fails if a required slot returns nil.

See [`examples/basic/cmd/server/user_admin.go`](examples/basic/cmd/server/user_admin.go) and [`superuser_field.go`](examples/basic/cmd/server/superuser_field.go) for a full field-policy override.

### Request helpers

Inside handlers and field code:

- `admin.GetUser(ctx)` — current authenticated user
- `admin.MustAdmin(ctx)` — access schema admin implementations
- `admin.UserHasPermission(ctx, user, name)` — RBAC check (`is_superuser` always succeeds)
- `vent.Forbidden` / `BadRequest` / … — client-safe HTTP errors

---

## CLI reference

```bash
# Scaffold auth schemas into ./ent/schema
go run github.com/troygilman/vent/cmd/vent init [--schema DIR] [--force]

# Generate Ent + Vent admin code
go run github.com/troygilman/vent/cmd/vent gen [--schema DIR] [--admin-path /admin/]
```

Alias: `generate` → `gen`.

---

## Example

A complete SQLite app lives in [`examples/basic/`](examples/basic/). From the repo root:

```bash
just gen   # Templ GUI + Vent admin codegen
just dev   # gen, then run the example server on :8080
```

Default login after first run: `admin@vent.com` / `test_user`.

Other recipes: `just migrations`, `just migrate`.

---

**Vent** — _Breathe life into your admin panels._
