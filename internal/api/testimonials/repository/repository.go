package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"sea-catering-backend/internal/api/testimonials"
	"sea-catering-backend/internal/entity"
)

type TestimonialRepository interface {
	Create(ctx context.Context, testimonial *entity.Testimonial) error
	GetByID(ctx context.Context, id string) (*entity.Testimonial, error)
	GetAll(ctx context.Context) ([]entity.Testimonial, error)
	GetApproved(ctx context.Context) ([]entity.Testimonial, error)
	GetPending(ctx context.Context) ([]entity.Testimonial, error)
	Update(ctx context.Context, testimonial *entity.Testimonial) error
	Delete(ctx context.Context, id string) error
	GetByRating(ctx context.Context, rating int) ([]entity.Testimonial, error)
	Count(ctx context.Context) (int, error)
	CountApproved(ctx context.Context) (int, error)
	CountPending(ctx context.Context) (int, error)
}

type testimonialRepository struct {
	db *sqlx.DB
}

func NewTestimonialRepository(db *sqlx.DB) TestimonialRepository {
	return &testimonialRepository{
		db: db,
	}
}

func (r *testimonialRepository) Create(ctx context.Context, testimonial *entity.Testimonial) error {
	query := `
		INSERT INTO testimonials (id, customer_name, message, rating, is_approved, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		testimonial.ID, testimonial.CustomerName, testimonial.Message,
		testimonial.Rating, testimonial.IsApproved,
		testimonial.CreatedAt, testimonial.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create testimonial: %w", err)
	}

	return nil
}

func (r *testimonialRepository) GetByID(ctx context.Context, id string) (*entity.Testimonial, error) {
	query := `
		SELECT id, customer_name, message, rating, is_approved, created_at, updated_at
		FROM testimonials
		WHERE id = $1
	`

	var testimonial entity.Testimonial
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&testimonial.ID, &testimonial.CustomerName, &testimonial.Message,
		&testimonial.Rating, &testimonial.IsApproved,
		&testimonial.CreatedAt, &testimonial.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, testimonials.ErrTestimonialNotFound
		}
		return nil, fmt.Errorf("failed to get testimonial: %w", err)
	}

	return &testimonial, nil
}

func (r *testimonialRepository) GetAll(ctx context.Context) ([]entity.Testimonial, error) {
	query := `
		SELECT id, customer_name, message, rating, is_approved, created_at, updated_at
		FROM testimonials
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all testimonials: %w", err)
	}
	defer rows.Close()

	var testimonialList []entity.Testimonial
	for rows.Next() {
		var testimonial entity.Testimonial
		err := rows.Scan(
			&testimonial.ID, &testimonial.CustomerName, &testimonial.Message,
			&testimonial.Rating, &testimonial.IsApproved,
			&testimonial.CreatedAt, &testimonial.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan testimonial: %w", err)
		}
		testimonialList = append(testimonialList, testimonial)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return testimonialList, nil
}

func (r *testimonialRepository) GetApproved(ctx context.Context) ([]entity.Testimonial, error) {
	query := `
		SELECT id, customer_name, message, rating, is_approved, created_at, updated_at
		FROM testimonials
		WHERE is_approved = true
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get approved testimonials: %w", err)
	}
	defer rows.Close()

	var testimonialList []entity.Testimonial
	for rows.Next() {
		var testimonial entity.Testimonial
		err := rows.Scan(
			&testimonial.ID, &testimonial.CustomerName, &testimonial.Message,
			&testimonial.Rating, &testimonial.IsApproved,
			&testimonial.CreatedAt, &testimonial.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan testimonial: %w", err)
		}
		testimonialList = append(testimonialList, testimonial)
	}

	return testimonialList, nil
}

func (r *testimonialRepository) GetPending(ctx context.Context) ([]entity.Testimonial, error) {
	query := `
		SELECT id, customer_name, message, rating, is_approved, created_at, updated_at
		FROM testimonials
		WHERE is_approved = false
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending testimonials: %w", err)
	}
	defer rows.Close()

	var testimonialList []entity.Testimonial
	for rows.Next() {
		var testimonial entity.Testimonial
		err := rows.Scan(
			&testimonial.ID, &testimonial.CustomerName, &testimonial.Message,
			&testimonial.Rating, &testimonial.IsApproved,
			&testimonial.CreatedAt, &testimonial.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan testimonial: %w", err)
		}
		testimonialList = append(testimonialList, testimonial)
	}

	return testimonialList, nil
}

func (r *testimonialRepository) Update(ctx context.Context, testimonial *entity.Testimonial) error {
	query := `
		UPDATE testimonials
		SET customer_name = $2, message = $3, rating = $4, is_approved = $5, updated_at = $6
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		testimonial.ID, testimonial.CustomerName, testimonial.Message,
		testimonial.Rating, testimonial.IsApproved, testimonial.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update testimonial: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return testimonials.ErrTestimonialNotFound
	}

	return nil
}

func (r *testimonialRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM testimonials WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete testimonial: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return testimonials.ErrTestimonialNotFound
	}

	return nil
}

func (r *testimonialRepository) GetByRating(ctx context.Context, rating int) ([]entity.Testimonial, error) {
	query := `
		SELECT id, customer_name, message, rating, is_approved, created_at, updated_at
		FROM testimonials
		WHERE rating = $1 AND is_approved = true
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, rating)
	if err != nil {
		return nil, fmt.Errorf("failed to get testimonials by rating: %w", err)
	}
	defer rows.Close()

	var testimonialList []entity.Testimonial
	for rows.Next() {
		var testimonial entity.Testimonial
		err := rows.Scan(
			&testimonial.ID, &testimonial.CustomerName, &testimonial.Message,
			&testimonial.Rating, &testimonial.IsApproved,
			&testimonial.CreatedAt, &testimonial.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan testimonial: %w", err)
		}
		testimonialList = append(testimonialList, testimonial)
	}

	return testimonialList, nil
}

func (r *testimonialRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM testimonials`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count testimonials: %w", err)
	}

	return count, nil
}

func (r *testimonialRepository) CountApproved(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM testimonials WHERE is_approved = true`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count approved testimonials: %w", err)
	}

	return count, nil
}

func (r *testimonialRepository) CountPending(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM testimonials WHERE is_approved = false`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count pending testimonials: %w", err)
	}

	return count, nil
}
