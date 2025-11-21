package main

import (
	"log"

	"vent"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

func main() {
	err := entc.Generate("./ent/schema",
		&gen.Config{},
		entc.Extensions(vent.NewAdminExtension()),
	)
	if err != nil {
		log.Fatal("running ent codegen:", err)
	}
}
