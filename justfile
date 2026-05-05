dev: gen
    go run ./examples/basic/cmd/server

gen:
    npx @tailwindcss/cli -i ./tailwind.css -o ./static/style.css
    templ generate ./templates/gui/
    go run ./cmd/vent gen --schema ./examples/basic/ent/schema

migrations:
    go run examples/basic/ent/migrate/main.go create_users

migrate:
    atlas migrate apply --dir "file://examples/basic/ent/migrate/migrations" --url "sqlite://tmp/test.db?_fk=1"
