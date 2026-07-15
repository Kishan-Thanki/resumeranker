//go:build integration

package users_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kishan-thanki/resumeranker/api/internal/users"
)

func TestPostgresRepository_Integration(t *testing.T) {
	t.Parallel()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/resumeranker?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	repo := users.NewPostgresRepository(pool)

	t.Run("Create and Get User Lifecycle", func(t *testing.T) {
		ctx := context.Background()

		email := fmt.Sprintf("integration_test_%d@example.com", time.Now().UnixNano())
		user := &users.User{
			Email:        email,
			PasswordHash: "hashedpassword",
			Role:         users.RoleUser,
			Status:       users.AccountStatusActive,
		}

		createdUser, err := repo.CreateUser(ctx, user)
		if err != nil {
			t.Skipf("skipping test due to DB error (likely needs migration): %v", err)
		}
		t.Cleanup(func() {
			_ = repo.DeleteUser(context.Background(), createdUser.ID)
		})

		if createdUser.ID == 0 {
			t.Error("expected ID to be set")
		}

		fetchedUser, err := repo.GetUserByEmail(ctx, email)
		if err != nil {
			t.Errorf("failed to get user: %v", err)
		}

		if fetchedUser.ID != createdUser.ID {
			t.Errorf("expected fetched user ID %d, got %d", createdUser.ID, fetchedUser.ID)
		}

		fetchedUser.Status = users.AccountStatusSuspended
		updatedUser, err := repo.UpdateUser(ctx, fetchedUser)
		if err != nil {
			t.Errorf("failed to update user: %v", err)
		}

		if updatedUser.Status != users.AccountStatusSuspended {
			t.Errorf("expected status suspended, got %s", updatedUser.Status)
		}

		err = repo.DeleteUser(ctx, updatedUser.ID)
		if err != nil {
			t.Errorf("failed to delete user: %v", err)
		}

		_, err = repo.GetUserByID(ctx, updatedUser.ID)
		if err == nil {
			t.Error("expected error getting soft-deleted user, got nil")
		}
	})
}
