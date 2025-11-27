package main

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"vent/ent"
	"vent/utils"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	client, err := ent.Open("sqlite3", "file:tmp/test.db?_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer client.Close()

	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	ctx := context.Background()

	passwordHash, err := utils.HashPassword("test_user")
	if err != nil {
		panic(err)
	}

	client.User.Create().
		SetEmail("admin@vent.com").
		SetPasswordHash(passwordHash).
		Save(ctx)

	for idx := range 1000 {
		client.Permission.Create().SetName(strconv.Itoa(idx)).Save(ctx)
	}

	mux := http.NewServeMux()
	mux.Handle("/admin/", ent.NewAdminHandler(ent.AdminConfig{
		Client: client,
		Secret: []byte("secret"),
	}))
	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
