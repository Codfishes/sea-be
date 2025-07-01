package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"sea-catering-backend/internal/api/auth"
	"sea-catering-backend/internal/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByPhone(ctx context.Context, phone string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
	UpdateProfileImage(ctx context.Context, userID uuid.UUID, imageURL string) error
	MarkEmailVerified(ctx context.Context, userID uuid.UUID) error
	MarkPhoneVerified(ctx context.Context, userID uuid.UUID) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByPhone(ctx context.Context, phone string) (bool, error)
	CountActiveUsers(ctx context.Context) (int, error)
	CountTotalUsers(ctx context.Context) (int, error)
}

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, name, email, phone, password, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Name, user.Email, user.Phone, user.Password,
		user.Role, user.IsActive, user.CreatedAt, user.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505":
				if pqErr.Constraint == "users_email_key" {
					return auth.ErrEmailAlreadyExists
				}
				if pqErr.Constraint == "users_phone_key" {
					return auth.ErrPhoneAlreadyExists
				}
			}
		}
		return err
	}

	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	query := `
		SELECT id, email, name, phone, password, is_verified, email_verified_at,
		       phone_verified_at, profile_image_url, role, is_active, last_login_at,
		       created_at, updated_at
		FROM users
		WHERE id = $1 AND is_active = true
	`

	var user entity.User
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, auth.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, email, name, phone, password, is_verified, email_verified_at,
		       phone_verified_at, profile_image_url, role, is_active, last_login_at,
		       created_at, updated_at
		FROM users
		WHERE email = $1 AND is_active = true
	`

	var user entity.User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, auth.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByPhone(ctx context.Context, phone string) (*entity.User, error) {
	query := `
		SELECT id, email, name, phone, password, is_verified, email_verified_at,
		       phone_verified_at, profile_image_url, role, is_active, last_login_at,
		       created_at, updated_at
		FROM users
		WHERE phone = $1 AND is_active = true
	`

	var user entity.User
	err := r.db.GetContext(ctx, &user, query, phone)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, auth.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET name = $2, phone = $3, updated_at = $4
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Name, user.Phone, time.Now(),
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505":
				if pqErr.Constraint == "users_phone_key" {
					return auth.ErrPhoneAlreadyExists
				}
			}
		}
		return err
	}

	return nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	query := `
		UPDATE users
		SET password = $2, updated_at = $3
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, userID, hashedPassword, time.Now())
	return err
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET last_login_at = $2, updated_at = $2
		WHERE id = $1
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, userID, now)
	return err
}

func (r *userRepository) UpdateProfileImage(ctx context.Context, userID uuid.UUID, imageURL string) error {
	query := `
		UPDATE users
		SET profile_image_url = $2, updated_at = $3
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, userID, imageURL, time.Now())
	return err
}

func (r *userRepository) MarkEmailVerified(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET is_verified = true, email_verified_at = $2, updated_at = $2
		WHERE id = $1
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, userID, now)
	return err
}

func (r *userRepository) MarkPhoneVerified(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET phone_verified_at = $2, updated_at = $2
		WHERE id = $1
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, userID, now)
	return err
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, email)
	return exists, err
}

func (r *userRepository) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE phone = $1)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, phone)
	return exists, err
}

func (r *userRepository) CountActiveUsers(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM users WHERE is_active = true`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count active users: %w", err)
	}

	return count, nil
}

func (r *userRepository) CountTotalUsers(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM users`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count total users: %w", err)
	}

	return count, nil
}
