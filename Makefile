gen:
	npx @tailwindcss/cli -i ./tailwind.css -o ./static/style.css
	templ generate ./templates/gui/
	go run ./cmd/gen

server:
	go run ./cmd/server

migrations:
	go run ent/migrate/main.go create_users

migrate:
	atlas migrate apply --dir "file://ent/migrate/migrations" --url "sqlite://tmp/test.db?_fk=1"
