package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/troygilman/vent/auth"
	"github.com/troygilman/vent/examples/basic/ent"
	"github.com/troygilman/vent/examples/basic/ent/author"
	"github.com/troygilman/vent/examples/basic/ent/book"
	"github.com/troygilman/vent/examples/basic/ent/review"
	"github.com/troygilman/vent/examples/basic/ent/user"
)

const (
	seedPassword        = "test_user"
	seedBulkUserCount   = 500
	seedBulkAuthorCount = 200
	seedBulkBookCount   = 2500
	seedBulkReviewCount = 4000
	seedBulkBatchSize   = 200
)

var (
	bookAdjectives = []string{
		"Silent", "Hidden", "Ancient", "Digital", "Forgotten", "Crystal",
		"Northern", "Iron", "Velvet", "Hollow", "Solar", "Midnight",
	}
	bookNouns = []string{
		"Library", "Engine", "Garden", "Ledger", "Compass", "Archive",
		"Machine", "Harbor", "Cipher", "Atlas", "Workshop", "Orchard",
	}
	reviewBodies = []string{
		"Could not put it down.",
		"Uneven, but the middle third is excellent.",
		"A solid reference more than a story.",
		"Worth reading twice.",
		"Useful, but still a draft.",
	}
)

func seedAdminUser(ctx context.Context, client *ent.Client, credentialGenerator auth.CredentialGenerator) error {
	exists, err := client.User.Query().Where(user.EmailEQ("admin@vent.com")).Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	passwordHash, err := credentialGenerator.Generate(seedPassword)
	if err != nil {
		return err
	}

	_, err = client.User.Create().
		SetEmail("admin@vent.com").
		SetPasswordHash(passwordHash).
		SetIsStaff(true).
		SetIsSuperuser(true).
		Save(ctx)
	return err
}

func seedUser(ctx context.Context, client *ent.Client, credentialGenerator auth.CredentialGenerator, email, password string, staff bool) (*ent.User, error) {
	existing, err := client.User.Query().Where(user.EmailEQ(email)).Only(ctx)
	if err == nil {
		return existing, nil
	}
	if !ent.IsNotFound(err) {
		return nil, err
	}

	passwordHash, err := credentialGenerator.Generate(password)
	if err != nil {
		return nil, err
	}
	return client.User.Create().
		SetEmail(email).
		SetPasswordHash(passwordHash).
		SetIsStaff(staff).
		Save(ctx)
}

func seedDemoData(ctx context.Context, client *ent.Client, credentialGenerator auth.CredentialGenerator) error {
	if count, err := client.Book.Query().Count(ctx); err != nil {
		return err
	} else if count >= seedBulkBookCount {
		log.Printf("demo data already present (%d books), skipping seed", count)
		return nil
	}

	passwordHash, err := credentialGenerator.Generate(seedPassword)
	if err != nil {
		return err
	}

	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	txClient := tx.Client()

	namedUsers, namedAuthors, err := seedNamedShowcase(ctx, txClient, credentialGenerator)
	if err != nil {
		return err
	}

	users, err := seedBulkUsers(ctx, txClient, passwordHash)
	if err != nil {
		return err
	}
	users = append(namedUsers, users...)

	authors, err := seedBulkAuthors(ctx, txClient, users[len(namedUsers):])
	if err != nil {
		return err
	}
	authors = append(namedAuthors, authors...)

	books, err := seedBulkBooks(ctx, txClient, authors)
	if err != nil {
		return err
	}

	if err = seedBulkReviews(ctx, txClient, books, users); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	log.Printf(
		"seeded demo data: %d users, %d authors, %d books, %d reviews",
		len(users)+1, // +1 for admin@vent.com, created separately
		len(authors),
		len(books)+2, // +2 named showcase books
		seedBulkReviewCount+2,
	)
	return nil
}

func seedNamedShowcase(ctx context.Context, client *ent.Client, credentialGenerator auth.CredentialGenerator) ([]*ent.User, []*ent.Author, error) {
	ada, err := seedUser(ctx, client, credentialGenerator, "ada@vent.com", seedPassword, true)
	if err != nil {
		return nil, nil, err
	}
	charles, err := seedUser(ctx, client, credentialGenerator, "charles@vent.com", seedPassword, true)
	if err != nil {
		return nil, nil, err
	}
	casey, err := seedUser(ctx, client, credentialGenerator, "casey@vent.com", seedPassword, false)
	if err != nil {
		return nil, nil, err
	}
	riley, err := seedUser(ctx, client, credentialGenerator, "riley@vent.com", seedPassword, false)
	if err != nil {
		return nil, nil, err
	}

	author, err := seedAuthorForUser(ctx, client, ada, true)
	if err != nil {
		return nil, nil, err
	}
	inactiveAuthor, err := seedAuthorForUser(ctx, client, charles, false)
	if err != nil {
		return nil, nil, err
	}

	publishedAt := time.Date(2024, 5, 1, 12, 0, 0, 0, time.UTC)
	showcaseBook, err := seedBookByTitle(ctx, client, "Analytical Engines", func() *ent.BookCreate {
		return client.Book.Create().
			SetTitle("Analytical Engines").
			SetPages(320).
			SetPublished(true).
			SetPublishedAt(publishedAt).
			SetInternalNotes("Feature showcase seed book").
			SetAuthor(author)
	})
	if err != nil {
		return nil, nil, err
	}

	_, err = seedBookByTitle(ctx, client, "Notes on the Difference Engine", func() *ent.BookCreate {
		return client.Book.Create().
			SetTitle("Notes on the Difference Engine").
			SetPages(48).
			SetPublished(false).
			SetAuthor(inactiveAuthor)
	})
	if err != nil {
		return nil, nil, err
	}

	if err = seedShowcaseReview(ctx, client, casey, showcaseBook, 5, "A clear tour of Vent's schema features."); err != nil {
		return nil, nil, err
	}
	if err = seedShowcaseReview(ctx, client, riley, showcaseBook, 3, "Useful, but still a draft."); err != nil {
		return nil, nil, err
	}

	return []*ent.User{ada, charles, casey, riley}, []*ent.Author{author, inactiveAuthor}, nil
}

func seedAuthorForUser(ctx context.Context, client *ent.Client, u *ent.User, active bool) (*ent.Author, error) {
	existing, err := client.Author.Query().Where(author.UserIDEQ(u.ID)).Only(ctx)
	if err == nil {
		return existing, nil
	}
	if !ent.IsNotFound(err) {
		return nil, err
	}
	return client.Author.Create().
		SetUser(u).
		SetActive(active).
		Save(ctx)
}

func seedBookByTitle(ctx context.Context, client *ent.Client, title string, create func() *ent.BookCreate) (*ent.Book, error) {
	existing, err := client.Book.Query().Where(book.TitleEQ(title)).Only(ctx)
	if err == nil {
		return existing, nil
	}
	if !ent.IsNotFound(err) {
		return nil, err
	}
	return create().Save(ctx)
}

func seedShowcaseReview(ctx context.Context, client *ent.Client, u *ent.User, b *ent.Book, rating int, body string) error {
	exists, err := client.Review.Query().Where(
		review.HasUserWith(user.IDEQ(u.ID)),
		review.HasBookWith(book.IDEQ(b.ID)),
	).Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = client.Review.Create().
		SetUser(u).
		SetRating(rating).
		SetBody(body).
		SetBook(b).
		Save(ctx)
	return err
}

func seedBulkUsers(ctx context.Context, client *ent.Client, passwordHash string) ([]*ent.User, error) {
	builders := make([]*ent.UserCreate, 0, seedBulkUserCount)
	for i := 1; i <= seedBulkUserCount; i++ {
		create := client.User.Create().
			SetEmail(fmt.Sprintf("user%04d@vent.com", i)).
			SetPasswordHash(passwordHash).
			SetIsStaff(i%10 == 0).
			SetIsActive(i%15 != 0)
		if i%4 != 0 {
			create.SetLastLogin(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i))
		}
		builders = append(builders, create)
	}
	return saveInBatches(builders, seedBulkBatchSize, func(batch []*ent.UserCreate) ([]*ent.User, error) {
		return client.User.CreateBulk(batch...).Save(ctx)
	})
}

func seedBulkAuthors(ctx context.Context, client *ent.Client, users []*ent.User) ([]*ent.Author, error) {
	count := seedBulkAuthorCount
	if count > len(users) {
		count = len(users)
	}
	builders := make([]*ent.AuthorCreate, 0, count)
	for i := 0; i < count; i++ {
		builders = append(builders, client.Author.Create().
			SetUser(users[i]).
			SetActive(i%7 != 0))
	}
	return saveInBatches(builders, seedBulkBatchSize, func(batch []*ent.AuthorCreate) ([]*ent.Author, error) {
		return client.Author.CreateBulk(batch...).Save(ctx)
	})
}

func seedBulkBooks(ctx context.Context, client *ent.Client, authors []*ent.Author) ([]*ent.Book, error) {
	builders := make([]*ent.BookCreate, 0, seedBulkBookCount)
	for i := 1; i <= seedBulkBookCount; i++ {
		published := i%3 != 0
		create := client.Book.Create().
			SetTitle(fmt.Sprintf("%s %s %04d", bookAdjectives[(i-1)%len(bookAdjectives)], bookNouns[(i-1)%len(bookNouns)], i)).
			SetPages(50 + (i % 950)).
			SetPublished(published).
			SetAuthor(authors[(i-1)%len(authors)])
		if published {
			create.SetPublishedAt(time.Date(2015, 1, 15, 12, 0, 0, 0, time.UTC).AddDate(0, 0, i))
		}
		if i%11 == 0 {
			create.SetInternalNotes(fmt.Sprintf("Bulk seed notes for book %04d", i))
		}
		builders = append(builders, create)
	}
	return saveInBatches(builders, seedBulkBatchSize, func(batch []*ent.BookCreate) ([]*ent.Book, error) {
		return client.Book.CreateBulk(batch...).Save(ctx)
	})
}

func seedBulkReviews(ctx context.Context, client *ent.Client, books []*ent.Book, users []*ent.User) error {
	builders := make([]*ent.ReviewCreate, 0, seedBulkReviewCount)
	for i := 1; i <= seedBulkReviewCount; i++ {
		create := client.Review.Create().
			SetRating(1 + (i % 5)).
			SetBook(books[(i-1)%len(books)]).
			SetUser(users[(i-1)%len(users)])
		if i%6 != 0 {
			create.SetBody(reviewBodies[(i-1)%len(reviewBodies)])
		}
		builders = append(builders, create)
	}
	_, err := saveInBatches(builders, seedBulkBatchSize, func(batch []*ent.ReviewCreate) ([]*ent.Review, error) {
		return client.Review.CreateBulk(batch...).Save(ctx)
	})
	return err
}

func saveInBatches[B any, E any](items []B, size int, save func([]B) ([]E, error)) ([]E, error) {
	out := make([]E, 0, len(items))
	for i := 0; i < len(items); i += size {
		end := i + size
		if end > len(items) {
			end = len(items)
		}
		created, err := save(items[i:end])
		if err != nil {
			return nil, err
		}
		out = append(out, created...)
	}
	return out, nil
}
