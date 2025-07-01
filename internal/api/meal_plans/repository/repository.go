package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"sea-catering-backend/internal/api/meal_plans"
	"sea-catering-backend/internal/entity"
)

type MealPlanRepository interface {
	Create(ctx context.Context, mealPlan *entity.MealPlan) error
	GetByID(ctx context.Context, id string) (*entity.MealPlan, error)
	GetByName(ctx context.Context, name string) (*entity.MealPlan, error)
	Update(ctx context.Context, mealPlan *entity.MealPlan) error
	Delete(ctx context.Context, id string) error

	List(ctx context.Context, params meal_plans.MealPlanListRequest) ([]entity.MealPlan, *meal_plans.PaginationMeta, error)
	GetActive(ctx context.Context) ([]entity.MealPlan, error)
	Search(ctx context.Context, query string, limit int) ([]entity.MealPlan, error)

	ExistsByName(ctx context.Context, name string, excludeID string) (bool, error)
	ExistsByID(ctx context.Context, id string) (bool, error)
	IsActive(ctx context.Context, id string) (bool, error)

	GetStats(ctx context.Context) (*meal_plans.MealPlanStatsResponse, error)
	GetSubscriptionCount(ctx context.Context, mealPlanID string) (int, error)
	GetPopularityRanking(ctx context.Context) ([]entity.MealPlan, error)

	UpdateActiveStatus(ctx context.Context, ids []string, isActive bool) error
	GetByIDs(ctx context.Context, ids []string) ([]entity.MealPlan, error)
}

type mealPlanRepository struct {
	db *sqlx.DB
}

func NewMealPlanRepository(db *sqlx.DB) MealPlanRepository {
	return &mealPlanRepository{
		db: db,
	}
}

func (r *mealPlanRepository) Create(ctx context.Context, mealPlan *entity.MealPlan) error {
	query := `
		INSERT INTO meal_plans (id, name, description, price, image_url, features, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		mealPlan.ID, mealPlan.Name, mealPlan.Description, mealPlan.Price,
		mealPlan.ImageURL, pq.Array(mealPlan.Features), mealPlan.IsActive,
		mealPlan.CreatedAt, mealPlan.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505":
				if strings.Contains(pqErr.Message, "name") {
					return meal_plans.ErrMealPlanNameTaken
				}
			}
		}
		return fmt.Errorf("failed to create meal plan: %w", err)
	}

	return nil
}

func (r *mealPlanRepository) GetByID(ctx context.Context, id string) (*entity.MealPlan, error) {
	query := `
		SELECT id, name, description, price, image_url, features, is_active, created_at, updated_at
		FROM meal_plans
		WHERE id = $1
	`

	var mealPlan entity.MealPlan
	var features pq.StringArray

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&mealPlan.ID, &mealPlan.Name, &mealPlan.Description, &mealPlan.Price,
		&mealPlan.ImageURL, &features, &mealPlan.IsActive,
		&mealPlan.CreatedAt, &mealPlan.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, meal_plans.ErrMealPlanNotFound
		}
		return nil, fmt.Errorf("failed to get meal plan: %w", err)
	}

	mealPlan.Features = []string(features)
	return &mealPlan, nil
}

func (r *mealPlanRepository) GetByName(ctx context.Context, name string) (*entity.MealPlan, error) {
	query := `
		SELECT id, name, description, price, image_url, features, is_active, created_at, updated_at
		FROM meal_plans
		WHERE LOWER(name) = LOWER($1)
	`

	var mealPlan entity.MealPlan
	var features pq.StringArray

	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&mealPlan.ID, &mealPlan.Name, &mealPlan.Description, &mealPlan.Price,
		&mealPlan.ImageURL, &features, &mealPlan.IsActive,
		&mealPlan.CreatedAt, &mealPlan.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, meal_plans.ErrMealPlanNotFound
		}
		return nil, fmt.Errorf("failed to get meal plan by name: %w", err)
	}

	mealPlan.Features = []string(features)
	return &mealPlan, nil
}

func (r *mealPlanRepository) Update(ctx context.Context, mealPlan *entity.MealPlan) error {
	query := `
		UPDATE meal_plans
		SET name = $2, description = $3, price = $4, image_url = $5, 
		    features = $6, is_active = $7, updated_at = $8
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		mealPlan.ID, mealPlan.Name, mealPlan.Description, mealPlan.Price,
		mealPlan.ImageURL, pq.Array(mealPlan.Features), mealPlan.IsActive,
		mealPlan.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505":
				if strings.Contains(pqErr.Message, "name") {
					return meal_plans.ErrMealPlanNameTaken
				}
			}
		}
		return fmt.Errorf("failed to update meal plan: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return meal_plans.ErrMealPlanNotFound
	}

	return nil
}

func (r *mealPlanRepository) Delete(ctx context.Context, id string) error {

	subscriptionCount, err := r.GetSubscriptionCount(ctx, id)
	if err != nil {
		return err
	}

	if subscriptionCount > 0 {
		return meal_plans.ErrMealPlanHasSubscriptions
	}

	query := `DELETE FROM meal_plans WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete meal plan: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return meal_plans.ErrMealPlanNotFound
	}

	return nil
}

func (r *mealPlanRepository) List(ctx context.Context, params meal_plans.MealPlanListRequest) ([]entity.MealPlan, *meal_plans.PaginationMeta, error) {

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.SortBy == "" {
		params.SortBy = "created_at"
	}
	if params.SortDir == "" {
		params.SortDir = "desc"
	}

	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if params.IsActive != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *params.IsActive)
		argIndex++
	}

	if params.Search != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("(LOWER(name) LIKE LOWER($%d) OR LOWER(description) LIKE LOWER($%d))", argIndex, argIndex))
		args = append(args, "%"+params.Search+"%")
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM meal_plans %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count meal plans: %w", err)
	}

	offset := (params.Page - 1) * params.Limit
	totalPages := (total + params.Limit - 1) / params.Limit

	validSortFields := map[string]bool{
		"name":       true,
		"price":      true,
		"created_at": true,
		"updated_at": true,
	}

	if !validSortFields[params.SortBy] {
		return nil, nil, meal_plans.ErrInvalidSortField
	}

	orderClause := fmt.Sprintf("ORDER BY %s %s", params.SortBy, strings.ToUpper(params.SortDir))

	query := fmt.Sprintf(`
		SELECT id, name, description, price, image_url, features, is_active, created_at, updated_at
		FROM meal_plans
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderClause, argIndex, argIndex+1)

	args = append(args, params.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list meal plans: %w", err)
	}
	defer rows.Close()

	var mealPlans []entity.MealPlan
	for rows.Next() {
		var mealPlan entity.MealPlan
		var features pq.StringArray

		err := rows.Scan(
			&mealPlan.ID, &mealPlan.Name, &mealPlan.Description, &mealPlan.Price,
			&mealPlan.ImageURL, &features, &mealPlan.IsActive,
			&mealPlan.CreatedAt, &mealPlan.UpdatedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan meal plan: %w", err)
		}

		mealPlan.Features = []string(features)
		mealPlans = append(mealPlans, mealPlan)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("row iteration error: %w", err)
	}

	meta := &meal_plans.PaginationMeta{
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    params.Page < totalPages,
		HasPrev:    params.Page > 1,
	}

	return mealPlans, meta, nil
}

func (r *mealPlanRepository) GetActive(ctx context.Context) ([]entity.MealPlan, error) {
	query := `
		SELECT id, name, description, price, image_url, features, is_active, created_at, updated_at
		FROM meal_plans
		WHERE is_active = true
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active meal plans: %w", err)
	}
	defer rows.Close()

	var mealPlans []entity.MealPlan
	for rows.Next() {
		var mealPlan entity.MealPlan
		var features pq.StringArray

		err := rows.Scan(
			&mealPlan.ID, &mealPlan.Name, &mealPlan.Description, &mealPlan.Price,
			&mealPlan.ImageURL, &features, &mealPlan.IsActive,
			&mealPlan.CreatedAt, &mealPlan.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan meal plan: %w", err)
		}

		mealPlan.Features = []string(features)
		mealPlans = append(mealPlans, mealPlan)
	}

	return mealPlans, nil
}

func (r *mealPlanRepository) Search(ctx context.Context, query string, limit int) ([]entity.MealPlan, error) {
	if limit <= 0 {
		limit = 10
	}

	searchQuery := `
		SELECT id, name, description, price, image_url, features, is_active, created_at, updated_at
		FROM meal_plans
		WHERE is_active = true AND (
			LOWER(name) LIKE LOWER($1) OR 
			LOWER(description) LIKE LOWER($1) OR
			EXISTS (
				SELECT 1 FROM unnest(features) AS feature 
				WHERE LOWER(feature) LIKE LOWER($1)
			)
		)
		ORDER BY 
			CASE WHEN LOWER(name) LIKE LOWER($1) THEN 1 ELSE 2 END,
			name
		LIMIT $2
	`

	searchPattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, searchQuery, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search meal plans: %w", err)
	}
	defer rows.Close()

	var mealPlans []entity.MealPlan
	for rows.Next() {
		var mealPlan entity.MealPlan
		var features pq.StringArray

		err := rows.Scan(
			&mealPlan.ID, &mealPlan.Name, &mealPlan.Description, &mealPlan.Price,
			&mealPlan.ImageURL, &features, &mealPlan.IsActive,
			&mealPlan.CreatedAt, &mealPlan.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan meal plan: %w", err)
		}

		mealPlan.Features = []string(features)
		mealPlans = append(mealPlans, mealPlan)
	}

	return mealPlans, nil
}

func (r *mealPlanRepository) ExistsByName(ctx context.Context, name string, excludeID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM meal_plans 
			WHERE LOWER(name) = LOWER($1) AND ($2 = '' OR id != $2)
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, name, excludeID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if meal plan exists: %w", err)
	}

	return exists, nil
}

func (r *mealPlanRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM meal_plans WHERE id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if meal plan exists: %w", err)
	}

	return exists, nil
}

func (r *mealPlanRepository) IsActive(ctx context.Context, id string) (bool, error) {
	query := `SELECT is_active FROM meal_plans WHERE id = $1`

	var isActive bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&isActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, meal_plans.ErrMealPlanNotFound
		}
		return false, fmt.Errorf("failed to check meal plan status: %w", err)
	}

	return isActive, nil
}

func (r *mealPlanRepository) GetStats(ctx context.Context) (*meal_plans.MealPlanStatsResponse, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active,
			COUNT(CASE WHEN is_active = false THEN 1 END) as inactive,
			COALESCE(AVG(price), 0) as average_price
		FROM meal_plans
	`

	var stats meal_plans.MealPlanStatsResponse
	err := r.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalMealPlans,
		&stats.ActiveMealPlans,
		&stats.InactiveMealPlans,
		&stats.AveragePrice,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get meal plan stats: %w", err)
	}

	popularQuery := `
		SELECT mp.name
		FROM meal_plans mp
		LEFT JOIN subscriptions s ON mp.id = s.meal_plan_id AND s.status = 'active'
		WHERE mp.is_active = true
		GROUP BY mp.id, mp.name
		ORDER BY COUNT(s.id) DESC
		LIMIT 1
	`

	err = r.db.QueryRowContext(ctx, popularQuery).Scan(&stats.MostPopularPlan)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get most popular plan: %w", err)
	}

	subscriberQuery := `
		SELECT COUNT(DISTINCT user_id)
		FROM subscriptions s
		JOIN meal_plans mp ON s.meal_plan_id = mp.id
		WHERE s.status = 'active' AND mp.is_active = true
	`

	err = r.db.QueryRowContext(ctx, subscriberQuery).Scan(&stats.TotalSubscribers)
	if err != nil {
		return nil, fmt.Errorf("failed to get total subscribers: %w", err)
	}

	return &stats, nil
}

func (r *mealPlanRepository) GetSubscriptionCount(ctx context.Context, mealPlanID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM subscriptions
		WHERE meal_plan_id = $1 AND status = 'active'
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, mealPlanID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get subscription count: %w", err)
	}

	return count, nil
}

func (r *mealPlanRepository) GetPopularityRanking(ctx context.Context) ([]entity.MealPlan, error) {
	query := `
		SELECT 
			mp.id, mp.name, mp.description, mp.price, mp.image_url, 
			mp.features, mp.is_active, mp.created_at, mp.updated_at,
			COUNT(s.id) as subscription_count
		FROM meal_plans mp
		LEFT JOIN subscriptions s ON mp.id = s.meal_plan_id AND s.status = 'active'
		WHERE mp.is_active = true
		GROUP BY mp.id, mp.name, mp.description, mp.price, mp.image_url, 
		         mp.features, mp.is_active, mp.created_at, mp.updated_at
		ORDER BY subscription_count DESC, mp.name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get popularity ranking: %w", err)
	}
	defer rows.Close()

	var mealPlans []entity.MealPlan
	for rows.Next() {
		var mealPlan entity.MealPlan
		var features pq.StringArray
		var subscriptionCount int

		err := rows.Scan(
			&mealPlan.ID, &mealPlan.Name, &mealPlan.Description, &mealPlan.Price,
			&mealPlan.ImageURL, &features, &mealPlan.IsActive,
			&mealPlan.CreatedAt, &mealPlan.UpdatedAt, &subscriptionCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan meal plan: %w", err)
		}

		mealPlan.Features = []string(features)
		mealPlans = append(mealPlans, mealPlan)
	}

	return mealPlans, nil
}

func (r *mealPlanRepository) UpdateActiveStatus(ctx context.Context, ids []string, isActive bool) error {
	if len(ids) == 0 {
		return nil
	}

	query := `
		UPDATE meal_plans 
		SET is_active = $1, updated_at = $2
		WHERE id = ANY($3)
	`

	_, err := r.db.ExecContext(ctx, query, isActive, time.Now(), pq.Array(ids))
	if err != nil {
		return fmt.Errorf("failed to update active status: %w", err)
	}

	return nil
}

func (r *mealPlanRepository) GetByIDs(ctx context.Context, ids []string) ([]entity.MealPlan, error) {
	if len(ids) == 0 {
		return []entity.MealPlan{}, nil
	}

	query := `
		SELECT id, name, description, price, image_url, features, is_active, created_at, updated_at
		FROM meal_plans
		WHERE id = ANY($1)
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(ids))
	if err != nil {
		return nil, fmt.Errorf("failed to get meal plans by IDs: %w", err)
	}
	defer rows.Close()

	var mealPlans []entity.MealPlan
	for rows.Next() {
		var mealPlan entity.MealPlan
		var features pq.StringArray

		err := rows.Scan(
			&mealPlan.ID, &mealPlan.Name, &mealPlan.Description, &mealPlan.Price,
			&mealPlan.ImageURL, &features, &mealPlan.IsActive,
			&mealPlan.CreatedAt, &mealPlan.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan meal plan: %w", err)
		}

		mealPlan.Features = []string(features)
		mealPlans = append(mealPlans, mealPlan)
	}

	return mealPlans, nil
}
