dev: gen
    go run ./cmd/server

gen:
    npx @tailwindcss/cli -i ./tailwind.css -o ./static/style.css
    templ generate ./templates/gui/
    go run ./cmd/gen

migrations:
    go run ent/migrate/main.go create_users

migrate:
    atlas migrate apply --dir "file://ent/migrate/migrations" --url "sqlite://tmp/test.db?_fk=1"
