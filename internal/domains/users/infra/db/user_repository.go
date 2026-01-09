package db

import (
	"context"
	"encoding/json"
	"time"

	"farohq-core-app/internal/domains/users/domain"
	"farohq-core-app/internal/domains/users/domain/model"
	"farohq-core-app/internal/domains/users/domain/ports/outbound"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository implements the outbound.UserRepository interface
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new PostgreSQL user repository
func NewUserRepository(db *pgxpool.Pool) outbound.UserRepository {
	return &UserRepository{
		db: db,
	}
}

// FindByClerkUserID finds a user by Clerk user ID
func (r *UserRepository) FindByClerkUserID(ctx context.Context, clerkUserID string) (*model.User, error) {
	query := `
		SELECT id, clerk_user_id, email, first_name, last_name, full_name, image_url, phone_numbers, created_at, updated_at, last_sign_in_at
		FROM users
		WHERE clerk_user_id = $1
	`

	var (
		id            string
		dbClerkUserID string
		email         *string
		firstName     *string
		lastName      *string
		fullName      *string
		imageURL      *string
		phoneNumbersJSON []byte
		createdAt     time.Time
		updatedAt     time.Time
		lastSignInAt  *time.Time
	)

	err := r.db.QueryRow(ctx, query, clerkUserID).Scan(
		&id,
		&dbClerkUserID,
		&email,
		&firstName,
		&lastName,
		&fullName,
		&imageURL,
		&phoneNumbersJSON,
		&createdAt,
		&updatedAt,
		&lastSignInAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	// Parse UUID
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	// Parse phone numbers JSON
	var phoneNumbers []string
	if len(phoneNumbersJSON) > 0 {
		if err := json.Unmarshal(phoneNumbersJSON, &phoneNumbers); err != nil {
			return nil, err
		}
	}

	return model.NewUserWithID(
		userID,
		dbClerkUserID,
		stringPtr(email),
		stringPtr(firstName),
		stringPtr(lastName),
		stringPtr(fullName),
		stringPtr(imageURL),
		phoneNumbers,
		createdAt,
		updatedAt,
		lastSignInAt,
	), nil
}

// Save creates a new user
func (r *UserRepository) Save(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (id, clerk_user_id, email, first_name, last_name, full_name, image_url, phone_numbers, created_at, updated_at, last_sign_in_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	phoneNumbersJSON, err := json.Marshal(user.PhoneNumbers())
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, query,
		user.ID(),
		user.ClerkUserID(),
		nullString(user.Email()),
		nullString(user.FirstName()),
		nullString(user.LastName()),
		nullString(user.FullName()),
		nullString(user.ImageURL()),
		phoneNumbersJSON,
		user.CreatedAt(),
		user.UpdatedAt(),
		user.LastSignInAt(),
	)

	return err
}

// Update updates an existing user
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET email = $2, first_name = $3, last_name = $4, full_name = $5, image_url = $6, phone_numbers = $7, updated_at = $8, last_sign_in_at = $9
		WHERE clerk_user_id = $1
	`

	phoneNumbersJSON, err := json.Marshal(user.PhoneNumbers())
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, query,
		user.ClerkUserID(),
		nullString(user.Email()),
		nullString(user.FirstName()),
		nullString(user.LastName()),
		nullString(user.FullName()),
		nullString(user.ImageURL()),
		phoneNumbersJSON,
		user.UpdatedAt(),
		user.LastSignInAt(),
	)

	return err
}

// Helper functions
func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func stringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
