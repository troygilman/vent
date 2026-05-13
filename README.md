# Vent

**A lightweight, type-safe admin panel & CRUD framework for Go + Ent + Templ + HTMX**

Vent generates beautiful, fully functional admin interfaces from your Ent schemas with almost zero boilerplate.

---

## Features

- **Zero-config admin UI** — Just add an Ent extension and get a full admin panel
- **Built with modern tools**:
  - [Ent](https://entgo.io/) for ORM
  - [Templ](https://templ.guide/) for type-safe components
  - HTMX + [Datastar](https://datastar.dev/) for reactive frontend
  - Tailwind CSS for styling
- **Built-in auth system** (users, groups, permissions with RBAC)
- **Automatic CRUD** for any Ent schema
- **Highly customizable** via schema annotations
- **Lightweight & embeddable** — perfect for internal tools and admin panels

---

## Quick Start

### 1. Install

```bash
go get github.com/troygilman/vent
```

### 2. Add the Ent Extension

In your `ent/entc.go` (or wherever you configure codegen):

```go
package ent

import (
    "github.com/troygilman/vent"
)

func init() {
    entc.GenSchema = entc.Schema{
        Extensions: []entc.Extension{
            vent.NewAdminExtension(),
        },
    }
}
```

### 3. Use Auth Mixins (optional but recommended)

```go
// schema/authuser.go
type AuthUser struct {
    ent.Schema
}

func (AuthUser) Mixin() []ent.Mixin {
    return []ent.Mixin{
        vent.AuthUserMixin{GroupSchemaType: AuthGroup{}},
    }
}
```

(Do the same for `AuthGroup` and `AuthPermission`.)

### 4. Generate & Run

```bash
just gen          # or run the commands manually
go run ./examples/basic/cmd/server
```

Then visit `http://localhost:8080/admin/`

---

## Example

See [`examples/basic/`](examples/basic/) for a complete working example with SQLite.

---

## Development

```bash
just dev          # Generate + run example server
just gen          # Regenerate CSS, Templ, and Vent admin code
```

---

## Tech Stack

| Layer           | Technology                          |
|-----------------|-------------------------------------|
| Backend         | Go 1.26 + Ent                      |
| UI Components   | Templ + HTMX + Datastar            |
| Styling         | Tailwind CSS                       |
| Auth            | JWT + bcrypt                       |
| Database        | Any Ent-supported DB (SQLite demo) |

---

## Roadmap

- More form input types & validation
- Advanced filtering & search
- Dark mode
- Plugin system
- Export / import features

---

**Vent** — *Breathe life into your admin panels.*

Made with ❤️ by [Troy Gilman](https://github.com/troygilman)
