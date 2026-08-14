package main

import (
	"fmt"

	ent "github.com/troygilman/vent/examples/basic/ent"
	"github.com/troygilman/vent/examples/basic/ent/admin"
)

// ReviewAdmin labels reviews by the reviewing user's email.
type ReviewAdmin struct {
	admin.DefaultReviewAdmin
}

func (ReviewAdmin) Name(e *ent.Review) string {
	if e.Edges.User != nil {
		return e.Edges.User.Email
	}
	return fmt.Sprintf("%d", e.ID)
}
