gen:
	npx @tailwindcss/cli -i ./tailwind.css -o ./static/style.css
	templ generate ./templates/gui/
	go run ./cmd/gen

server:
	go run ./cmd/server
