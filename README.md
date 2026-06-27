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

    This writes `auth_user.go`, `auth_group.go`, and `auth_permission.go` into your schema directory. Each schema uses the corresponding Vent mixin:

    ```go
    func (AuthUser) Mixin() []ent.Mixin {
        return []ent.Mixin{
            vent.AuthUserMixin{GroupSchemaType: AuthGroup.Type},
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

## Example

A complete working example with SQLite lives in [`examples/basic/`](examples/basic/). Run it with:

```bash
just dev   # generate + run the example server
just gen   # regenerate Templ, CSS, and Vent admin code
```

---

**Vent** — _Breathe life into your admin panels._
