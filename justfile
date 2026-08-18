dev: gen
    go run ./examples/basic/cmd/server

gen:
    templ generate ./templates/gui/
    go run ./cmd/vent gen --schema ./examples/basic/ent/schema
    # CSS is now native (static/css/style.css). No Tailwind build step required.

migrations:
    go run examples/basic/ent/migrate/main.go create_users

migrate:
    atlas migrate apply --dir "file://examples/basic/ent/migrate/migrations" --url "sqlite://tmp/test.db?_fk=1"

docker-build:
    docker build -t vent-example .

docker: docker-build
    docker run --rm -p 8080:8080 vent-example

tail:
    sudo tailscale serve --bg http://localhost:8080
