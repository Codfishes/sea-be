package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"sea-catering-backend/internal/api/admin"
	"sea-catering-backend/internal/entity"
)

type AdminRepository interface {
	GetByID(ctx context.Context, id string) (*entity.AdminUser, error)
	GetByEmail(ctx context.Context, email string) (*entity.AdminUser, error)
	Create(ctx context.Context, admin *entity.AdminUser) error
	Update(ctx context.Context, admin *entity.AdminUser) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]entity.AdminUser, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	GetAllUsers(ctx context.Context, req admin.UserListRequest) ([]admin.UserResponse, *admin.PaginationMeta, error)
	GetUserByID(ctx context.Context, userID string) (*admin.UserResponse, error)
	UpdateUserStatus(ctx context.Context, userID string, isActive bool, reason string) error
	DeleteUser(ctx context.Context, userID string) error
	GetUserStats(ctx context.Context, userID string) (subscriptionCount int, totalSpent float64, error error)

	SearchSubscriptions(ctx context.Context, req admin.SubscriptionSearchRequest) ([]admin.SubscriptionSearchResponse, *admin.PaginationMeta, error)
	ForceCancelSubscription(ctx context.Context, subscriptionID, reason, adminComments string) error
	GetSubscriptionForCancel(ctx context.Context, subscriptionID string) (*admin.SubscriptionSearchResponse, error)
}

type adminRepository struct {
	db *sqlx.DB
}

func NewAdminRepository(db *sqlx.DB) AdminRepository {
	return &adminRepository{
		db: db,
	}
}

func (r *adminRepository) GetByID(ctx context.Context, id string) (*entity.AdminUser, error) {
	query := `
		SELECT id, email, name, password, role, created_at, updated_at
		FROM admin_users
		WHERE id = $1
	`

	var adminUser entity.AdminUser
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&adminUser.ID, &adminUser.Email, &adminUser.Name,
		&adminUser.Password, &adminUser.Role,
		&adminUser.CreatedAt, &adminUser.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, admin.ErrAdminNotFound
		}
		return nil, fmt.Errorf("failed to get admin by ID: %w", err)
	}

	return &adminUser, nil
}

func (r *adminRepository) GetByEmail(ctx context.Context, email string) (*entity.AdminUser, error) {
	query := `
		SELECT id, email, name, password, role, created_at, updated_at
		FROM admin_users
		WHERE email = $1
	`

	var adminUser entity.AdminUser
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&adminUser.ID, &adminUser.Email, &adminUser.Name,
		&adminUser.Password, &adminUser.Role,
		&adminUser.CreatedAt, &adminUser.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, admin.ErrAdminNotFound
		}
		return nil, fmt.Errorf("failed to get admin by email: %w", err)
	}

	return &adminUser, nil
}

func (r *adminRepository) Create(ctx context.Context, adminUser *entity.AdminUser) error {
	query := `
		INSERT INTO admin_users (id, email, name, password, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		adminUser.ID, adminUser.Email, adminUser.Name,
		adminUser.Password, adminUser.Role,
		adminUser.CreatedAt, adminUser.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	return nil
}

func (r *adminRepository) Update(ctx context.Context, adminUser *entity.AdminUser) error {
	query := `
		UPDATE admin_users
		SET email = $2, name = $3, password = $4, role = $5, updated_at = $6
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		adminUser.ID, adminUser.Email, adminUser.Name,
		adminUser.Password, adminUser.Role, adminUser.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update admin user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return admin.ErrAdminNotFound
	}

	return nil
}

func (r *adminRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM admin_users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete admin user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return admin.ErrAdminNotFound
	}

	return nil
}

func (r *adminRepository) List(ctx context.Context) ([]entity.AdminUser, error) {
	query := `
		SELECT id, email, name, password, role, created_at, updated_at
		FROM admin_users
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list admin users: %w", err)
	}
	defer rows.Close()

	var adminUsers []entity.AdminUser
	for rows.Next() {
		var adminUser entity.AdminUser
		err := rows.Scan(
			&adminUser.ID, &adminUser.Email, &adminUser.Name,
			&adminUser.Password, &adminUser.Role,
			&adminUser.CreatedAt, &adminUser.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan admin user: %w", err)
		}
		adminUsers = append(adminUsers, adminUser)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return adminUsers, nil
}

func (r *adminRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM admin_users WHERE email = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check admin existence: %w", err)
	}

	return exists, nil
}

func (r *adminRepository) GetAllUsers(ctx context.Context, req admin.UserListRequest) ([]admin.UserResponse, *admin.PaginationMeta, error) {

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortDir == "" {
		req.SortDir = "desc"
	}

	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Status == "active" {
		whereConditions = append(whereConditions, fmt.Sprintf("u.is_active = $%d", argIndex))
		args = append(args, true)
		argIndex++
	} else if req.Status == "inactive" {
		whereConditions = append(whereConditions, fmt.Sprintf("u.is_active = $%d", argIndex))
		args = append(args, false)
		argIndex++
	}

	if req.Role != "" && req.Role != "all" {
		whereConditions = append(whereConditions, fmt.Sprintf("u.role = $%d", argIndex))
		args = append(args, req.Role)
		argIndex++
	}

	if req.Search != "" {
		searchCondition := fmt.Sprintf("(LOWER(u.name) LIKE LOWER($%d) OR LOWER(u.email) LIKE LOWER($%d) OR u.phone LIKE $%d)", argIndex, argIndex, argIndex)
		whereConditions = append(whereConditions, searchCondition)
		args = append(args, "%"+req.Search+"%")
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM users u 
		%s
	`, whereClause)

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count users: %w", err)
	}

	offset := (req.Page - 1) * req.Limit
	totalPages := (total + req.Limit - 1) / req.Limit

	validSortFields := map[string]bool{
		"name":          true,
		"email":         true,
		"created_at":    true,
		"last_login_at": true,
	}

	if !validSortFields[req.SortBy] {
		req.SortBy = "created_at"
	}

	query := fmt.Sprintf(`
		SELECT 
			u.id, u.email, u.name, u.phone, u.is_verified, 
			u.email_verified_at, u.phone_verified_at, u.profile_image_url,
			u.role, u.is_active, u.last_login_at, u.created_at, u.updated_at,
			COALESCE(s.subscription_count, 0) as subscription_count,
			COALESCE(s.total_spent, 0) as total_spent
		FROM users u
		LEFT JOIN (
			SELECT 
				user_id,
				COUNT(*) as subscription_count,
				SUM(total_price) as total_spent
			FROM subscriptions 
			WHERE status != 'cancelled'
			GROUP BY user_id
		) s ON u.id = s.user_id
		%s
		ORDER BY u.%s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, req.SortBy, strings.ToUpper(req.SortDir), argIndex, argIndex+1)

	args = append(args, req.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []admin.UserResponse
	for rows.Next() {
		var user admin.UserResponse
		err := rows.Scan(
			&user.ID, &user.Email, &user.Name, &user.Phone, &user.IsVerified,
			&user.EmailVerifiedAt, &user.PhoneVerifiedAt, &user.ProfileImageURL,
			&user.Role, &user.IsActive, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
			&user.SubscriptionCount, &user.TotalSpent,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("row iteration error: %w", err)
	}

	meta := &admin.PaginationMeta{
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    req.Page < totalPages,
		HasPrev:    req.Page > 1,
	}

	return users, meta, nil
}

func (r *adminRepository) GetUserByID(ctx context.Context, userID string) (*admin.UserResponse, error) {
	query := `
		SELECT 
			u.id, u.email, u.name, u.phone, u.is_verified, 
			u.email_verified_at, u.phone_verified_at, u.profile_image_url,
			u.role, u.is_active, u.last_login_at, u.created_at, u.updated_at,
			COALESCE(s.subscription_count, 0) as subscription_count,
			COALESCE(s.total_spent, 0) as total_spent
		FROM users u
		LEFT JOIN (
			SELECT 
				user_id,
				COUNT(*) as subscription_count,
				SUM(total_price) as total_spent
			FROM subscriptions 
			WHERE status != 'cancelled'
			GROUP BY user_id
		) s ON u.id = s.user_id
		WHERE u.id = $1
	`

	var user admin.UserResponse
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.Email, &user.Name, &user.Phone, &user.IsVerified,
		&user.EmailVerifiedAt, &user.PhoneVerifiedAt, &user.ProfileImageURL,
		&user.Role, &user.IsActive, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
		&user.SubscriptionCount, &user.TotalSpent,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *adminRepository) UpdateUserStatus(ctx context.Context, userID string, isActive bool, reason string) error {
	query := `
		UPDATE users 
		SET is_active = $2, updated_at = $3
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, userID, isActive, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *adminRepository) DeleteUser(ctx context.Context, userID string) error {

	var activeSubscriptions int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM subscriptions WHERE user_id = $1 AND status IN ('active', 'paused')",
		userID).Scan(&activeSubscriptions)
	if err != nil {
		return fmt.Errorf("failed to check user subscriptions: %w", err)
	}

	if activeSubscriptions > 0 {
		return fmt.Errorf("cannot delete user with active subscriptions")
	}

	query := `
		UPDATE users 
		SET is_active = false, updated_at = $2, email = email || '_deleted_' || extract(epoch from now())
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, userID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *adminRepository) GetUserStats(ctx context.Context, userID string) (subscriptionCount int, totalSpent float64, error error) {
	query := `
		SELECT 
			COUNT(*) as subscription_count,
			COALESCE(SUM(total_price), 0) as total_spent
		FROM subscriptions 
		WHERE user_id = $1 AND status != 'cancelled'
	`

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&subscriptionCount, &totalSpent)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get user stats: %w", err)
	}

	return subscriptionCount, totalSpent, nil
}

func (r *adminRepository) SearchSubscriptions(ctx context.Context, req admin.SubscriptionSearchRequest) ([]admin.SubscriptionSearchResponse, *admin.PaginationMeta, error) {

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortDir == "" {
		req.SortDir = "desc"
	}

	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Status != "" && req.Status != "all" {
		whereConditions = append(whereConditions, fmt.Sprintf("s.status = $%d", argIndex))
		args = append(args, req.Status)
		argIndex++
	}

	if req.MealPlanID != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("s.meal_plan_id = $%d", argIndex))
		args = append(args, req.MealPlanID)
		argIndex++
	}

	if req.UserID != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("s.user_id = $%d", argIndex))
		args = append(args, req.UserID)
		argIndex++
	}

	if req.MinPrice > 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("s.total_price >= $%d", argIndex))
		args = append(args, req.MinPrice)
		argIndex++
	}

	if req.MaxPrice > 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("s.total_price <= $%d", argIndex))
		args = append(args, req.MaxPrice)
		argIndex++
	}

	if req.DateFrom != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("s.created_at >= $%d", argIndex))
		args = append(args, req.DateFrom)
		argIndex++
	}

	if req.DateTo != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("s.created_at <= $%d", argIndex))
		args = append(args, req.DateTo)
		argIndex++
	}

	if req.Search != "" {
		searchCondition := fmt.Sprintf(`(
			LOWER(u.name) LIKE LOWER($%d) OR 
			LOWER(u.email) LIKE LOWER($%d) OR 
			u.phone LIKE $%d OR
			LOWER(mp.name) LIKE LOWER($%d)
		)`, argIndex, argIndex, argIndex, argIndex)
		whereConditions = append(whereConditions, searchCondition)
		args = append(args, "%"+req.Search+"%")
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM subscriptions s
		JOIN users u ON s.user_id = u.id
		JOIN meal_plans mp ON s.meal_plan_id = mp.id
		%s
	`, whereClause)

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count subscriptions: %w", err)
	}

	offset := (req.Page - 1) * req.Limit
	totalPages := (total + req.Limit - 1) / req.Limit

	validSortFields := map[string]bool{
		"created_at":  true,
		"updated_at":  true,
		"total_price": true,
	}

	if !validSortFields[req.SortBy] {
		req.SortBy = "created_at"
	}

	query := fmt.Sprintf(`
		SELECT 
			s.id, s.user_id, u.name, u.email, u.phone,
			s.meal_plan_id, mp.name, s.meal_types, s.delivery_days,
			s.total_price, s.status, s.pause_start_date, s.pause_end_date,
			s.created_at, s.updated_at
		FROM subscriptions s
		JOIN users u ON s.user_id = u.id
		JOIN meal_plans mp ON s.meal_plan_id = mp.id
		%s
		ORDER BY s.%s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, req.SortBy, strings.ToUpper(req.SortDir), argIndex, argIndex+1)

	args = append(args, req.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []admin.SubscriptionSearchResponse
	for rows.Next() {
		var sub admin.SubscriptionSearchResponse
		var mealTypes, deliveryDays []string

		err := rows.Scan(
			&sub.ID, &sub.UserID, &sub.UserName, &sub.UserEmail, &sub.UserPhone,
			&sub.MealPlanID, &sub.MealPlanName, &mealTypes, &deliveryDays,
			&sub.TotalPrice, &sub.Status, &sub.PauseStart, &sub.PauseEnd,
			&sub.CreatedAt, &sub.UpdatedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan subscription: %w", err)
		}

		sub.MealTypes = mealTypes
		sub.DeliveryDays = deliveryDays
		subscriptions = append(subscriptions, sub)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("row iteration error: %w", err)
	}

	meta := &admin.PaginationMeta{
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    req.Page < totalPages,
		HasPrev:    req.Page > 1,
	}

	return subscriptions, meta, nil
}

func (r *adminRepository) ForceCancelSubscription(ctx context.Context, subscriptionID, reason, adminComments string) error {
	query := `
		UPDATE subscriptions 
		SET status = 'cancelled', 
		    pause_start_date = NULL, 
		    pause_end_date = NULL,
		    updated_at = $2
		WHERE id = $1 AND status != 'cancelled'
	`

	result, err := r.db.ExecContext(ctx, query, subscriptionID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("subscription not found or already cancelled")
	}

	return nil
}

func (r *adminRepository) GetSubscriptionForCancel(ctx context.Context, subscriptionID string) (*admin.SubscriptionSearchResponse, error) {
	query := `
		SELECT 
			s.id, s.user_id, u.name, u.email, u.phone,
			s.meal_plan_id, mp.name, s.meal_types, s.delivery_days,
			s.total_price, s.status, s.pause_start_date, s.pause_end_date,
			s.created_at, s.updated_at
		FROM subscriptions s
		JOIN users u ON s.user_id = u.id
		JOIN meal_plans mp ON s.meal_plan_id = mp.id
		WHERE s.id = $1
	`

	var sub admin.SubscriptionSearchResponse
	var mealTypes, deliveryDays []string

	err := r.db.QueryRowContext(ctx, query, subscriptionID).Scan(
		&sub.ID, &sub.UserID, &sub.UserName, &sub.UserEmail, &sub.UserPhone,
		&sub.MealPlanID, &sub.MealPlanName, &mealTypes, &deliveryDays,
		&sub.TotalPrice, &sub.Status, &sub.PauseStart, &sub.PauseEnd,
		&sub.CreatedAt, &sub.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("subscription not found")
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	sub.MealTypes = mealTypes
	sub.DeliveryDays = deliveryDays

	return &sub, nil
}
