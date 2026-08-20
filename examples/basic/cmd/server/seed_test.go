package main

import (
	"context"
	"testing"

	"github.com/troygilman/vent/auth"
	"github.com/troygilman/vent/examples/basic/ent/book"
	"github.com/troygilman/vent/examples/basic/ent/enttest"
	"github.com/troygilman/vent/examples/basic/ent/user"

	_ "github.com/mattn/go-sqlite3"
)

func TestSeedDemoDataCreatesThousandsOfRows(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:seed_test?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	ctx := context.Background()
	credentialGenerator := auth.NewBCryptCredentialGenerator()

	if err := seedAdminUser(ctx, client, credentialGenerator); err != nil {
		t.Fatalf("seed admin user: %v", err)
	}
	if err := seedDemoData(ctx, client, credentialGenerator); err != nil {
		t.Fatalf("seed demo data: %v", err)
	}

	userCount, err := client.User.Query().Count(ctx)
	if err != nil {
		t.Fatalf("count users: %v", err)
	}
	authorCount, err := client.Author.Query().Count(ctx)
	if err != nil {
		t.Fatalf("count authors: %v", err)
	}
	bookCount, err := client.Book.Query().Count(ctx)
	if err != nil {
		t.Fatalf("count books: %v", err)
	}
	reviewCount, err := client.Review.Query().Count(ctx)
	if err != nil {
		t.Fatalf("count reviews: %v", err)
	}

	wantUsers := 1 + 4 + seedBulkUserCount
	wantAuthors := 2 + seedBulkAuthorCount
	wantBooks := 2 + seedBulkBookCount
	wantReviews := 2 + seedBulkReviewCount

	if userCount != wantUsers {
		t.Errorf("users = %d, want %d", userCount, wantUsers)
	}
	if authorCount != wantAuthors {
		t.Errorf("authors = %d, want %d", authorCount, wantAuthors)
	}
	if bookCount != wantBooks {
		t.Errorf("books = %d, want %d", bookCount, wantBooks)
	}
	if reviewCount != wantReviews {
		t.Errorf("reviews = %d, want %d", reviewCount, wantReviews)
	}

	admin, err := client.User.Query().Where(user.EmailEQ("admin@vent.com")).Only(ctx)
	if err != nil {
		t.Fatalf("load admin user: %v", err)
	}
	if !admin.IsStaff || !admin.IsSuperuser {
		t.Errorf("admin@vent.com staff=%v superuser=%v, want both true", admin.IsStaff, admin.IsSuperuser)
	}

	inactiveUsers, err := client.User.Query().Where(user.IsActive(false)).Count(ctx)
	if err != nil {
		t.Fatalf("count inactive users: %v", err)
	}
	if inactiveUsers == 0 {
		t.Error("expected some inactive users for list-filter testing")
	}

	unpublished, err := client.Book.Query().Where(book.Published(false)).Count(ctx)
	if err != nil {
		t.Fatalf("count unpublished books: %v", err)
	}
	if unpublished == 0 {
		t.Error("expected some unpublished books for list-filter testing")
	}

	if err := seedDemoData(ctx, client, credentialGenerator); err != nil {
		t.Fatalf("second seed: %v", err)
	}
	bookCountAfter, err := client.Book.Query().Count(ctx)
	if err != nil {
		t.Fatalf("count books after second seed: %v", err)
	}
	if bookCountAfter != bookCount {
		t.Errorf("second seed changed book count from %d to %d", bookCount, bookCountAfter)
	}
}

func TestSeedDemoDataExpandsExistingShowcaseRows(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:seed_expand_test?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	ctx := context.Background()
	credentialGenerator := auth.NewBCryptCredentialGenerator()

	if err := seedAdminUser(ctx, client, credentialGenerator); err != nil {
		t.Fatalf("seed admin user: %v", err)
	}
	if _, _, err := seedNamedShowcase(ctx, client, credentialGenerator); err != nil {
		t.Fatalf("seed named showcase: %v", err)
	}

	beforeBooks, err := client.Book.Query().Count(ctx)
	if err != nil {
		t.Fatalf("count books before bulk seed: %v", err)
	}
	if beforeBooks != 2 {
		t.Fatalf("showcase books = %d, want 2", beforeBooks)
	}

	if err := seedDemoData(ctx, client, credentialGenerator); err != nil {
		t.Fatalf("expand seed: %v", err)
	}

	afterBooks, err := client.Book.Query().Count(ctx)
	if err != nil {
		t.Fatalf("count books after bulk seed: %v", err)
	}
	if afterBooks != beforeBooks+seedBulkBookCount {
		t.Errorf("books after expand = %d, want %d", afterBooks, beforeBooks+seedBulkBookCount)
	}

	userCount, err := client.User.Query().Count(ctx)
	if err != nil {
		t.Fatalf("count users after expand: %v", err)
	}
	if userCount != 1+4+seedBulkUserCount {
		t.Errorf("users after expand = %d, want %d", userCount, 1+4+seedBulkUserCount)
	}
}
