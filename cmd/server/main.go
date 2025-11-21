package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"vent/ent"
	"vent/utils"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	client, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer client.Close()
	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	ctx := context.Background()

	passwordHash, err := utils.HashPassword("test_user")
	if err != nil {
		panic(err)
	}

	u, err := client.User.Create().
		SetEmail("admin@vent.com").
		SetPasswordHash(passwordHash).
		Save(ctx)
	if err != nil {
		log.Println(fmt.Errorf("failed creating user: %w", err))
		return
	}
	log.Println("user was created: ", u)

	mux := http.NewServeMux()
	mux.Handle("/admin/", ent.NewAdminHandler(client, []byte("secret")))
	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
