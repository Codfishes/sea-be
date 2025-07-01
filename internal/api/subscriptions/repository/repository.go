package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sea-catering-backend/pkg/utils"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"sea-catering-backend/internal/entity"
	"sea-catering-backend/pkg/logger"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, subscription *entity.Subscription) error
	GetByID(ctx context.Context, id string) (*entity.SubscriptionWithDetails, error)
	GetByUserID(ctx context.Context, userID string) ([]entity.SubscriptionWithDetails, error)
	Update(ctx context.Context, subscription *entity.Subscription) error
	Delete(ctx context.Context, id string) error
	GetActiveSubscriptions(ctx context.Context) ([]entity.SubscriptionWithDetails, error)
	GetSubscriptionStats(ctx context.Context, startDate, endDate time.Time) (*SubscriptionStats, error)
	GetReactivationsCount(ctx context.Context, startDate, endDate time.Time) (int, error)
	ExistsByUserAndPlan(ctx context.Context, userID, planID string) (bool, error)
	GetExpiredSubscriptions(ctx context.Context) ([]entity.Subscription, error)
	BulkUpdateStatus(ctx context.Context, ids []string, status entity.SubscriptionStatus) error

	LogSubscriptionAction(ctx context.Context, subscriptionID, userID, action, oldStatus, newStatus string) error
	GetSubscriptionByIDForReactivation(ctx context.Context, id string) (*entity.Subscription, error)
}

type SubscriptionStats struct {
	TotalSubscriptions     int
	ActiveSubscriptions    int
	PausedSubscriptions    int
	CancelledSubscriptions int
	NewSubscriptions       int
	MonthlyRevenue         float64
	Reactivations          int
	SubscriptionsByPlan    map[string]int
}

type subscriptionRepository struct {
	db     *sqlx.DB
	logger *logger.Logger
	utils  *utils.Interface
}

func NewSubscriptionRepository(db *sqlx.DB, logger *logger.Logger, utils utils.Interface) SubscriptionRepository {
	return &subscriptionRepository{
		db:     db,
		logger: logger,
		utils:  &utils,
	}
}

func (r *subscriptionRepository) GetSubscriptionByIDForReactivation(ctx context.Context, id string) (*entity.Subscription, error) {
	query := `
        SELECT 
            id, user_id, meal_plan_id, meal_types, delivery_days,
            allergies, total_price, status, pause_start_date, pause_end_date,
            created_at, updated_at
        FROM subscriptions 
        WHERE id = $1
    `

	var sub entity.Subscription
	var mealTypes pq.StringArray
	var deliveryDays pq.StringArray

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&sub.ID, &sub.UserID, &sub.MealPlanID, &mealTypes, &deliveryDays,
		&sub.Allergies, &sub.TotalPrice, &sub.Status, &sub.PauseStartDate, &sub.PauseEndDate,
		&sub.CreatedAt, &sub.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("subscription not found")
		}
		r.logger.Error("Failed to get subscription for reactivation", logger.Fields{
			"error": err.Error(),
			"id":    id,
		})
		return nil, err
	}

	sub.MealTypes = make([]entity.MealType, len(mealTypes))
	for i, mt := range mealTypes {
		sub.MealTypes[i] = entity.MealType(mt)
	}

	sub.DeliveryDays = make([]entity.DeliveryDay, len(deliveryDays))
	for i, dd := range deliveryDays {
		sub.DeliveryDays[i] = entity.DeliveryDay(dd)
	}

	return &sub, nil
}

func (r *subscriptionRepository) LogSubscriptionAction(ctx context.Context, subscriptionID, userID, action, oldStatus, newStatus string) error {

	checkTableQuery := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'subscription_audit'
		)
	`

	var tableExists bool
	err := r.db.QueryRowContext(ctx, checkTableQuery).Scan(&tableExists)
	if err != nil || !tableExists {

		r.logger.Debug("Subscription audit table not found, skipping audit log", logger.Fields{
			"subscription_id": subscriptionID,
			"action":          action,
		})
		return nil
	}

	auditID := fmt.Sprintf("%s-%s", subscriptionID, time.Now().Format("20060102150405"))

	query := `
		INSERT INTO subscription_audit (
			id, subscription_id, user_id, old_status, new_status, 
			action, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = r.db.ExecContext(ctx, query,
		auditID, subscriptionID, userID, oldStatus, newStatus,
		action, time.Now(),
	)

	if err != nil {
		r.logger.Error("Failed to log subscription action", logger.Fields{
			"error":           err.Error(),
			"subscription_id": subscriptionID,
			"action":          action,
		})

	}

	return nil
}

func (r *subscriptionRepository) GetSubscriptionStats(ctx context.Context, startDate, endDate time.Time) (*SubscriptionStats, error) {
	query := `
        SELECT 
            COUNT(*) as total_subscriptions,
            COUNT(CASE WHEN status = 'active' THEN 1 END) as active_subscriptions,
            COUNT(CASE WHEN status = 'paused' THEN 1 END) as paused_subscriptions,
            COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled_subscriptions,
            COUNT(CASE WHEN created_at BETWEEN $1 AND $2 THEN 1 END) as new_subscriptions,
            COALESCE(SUM(CASE WHEN status = 'active' THEN total_price ELSE 0 END), 0) as monthly_revenue
        FROM subscriptions
    `

	var stats SubscriptionStats
	err := r.db.QueryRowContext(ctx, query, startDate, endDate).Scan(
		&stats.TotalSubscriptions,
		&stats.ActiveSubscriptions,
		&stats.PausedSubscriptions,
		&stats.CancelledSubscriptions,
		&stats.NewSubscriptions,
		&stats.MonthlyRevenue,
	)

	if err != nil {
		r.logger.Error("Failed to get subscription stats", logger.Fields{
			"error": err.Error(),
		})
		return nil, err
	}

	reactivations, err := r.GetReactivationsCount(ctx, startDate, endDate)
	if err != nil {
		r.logger.Warn("Failed to get reactivations count", logger.Fields{
			"error": err.Error(),
		})
		reactivations = 0
	}
	stats.Reactivations = reactivations

	planQuery := `
        SELECT mp.name, COUNT(*) as count
        FROM subscriptions s
        JOIN meal_plans mp ON s.meal_plan_id = mp.id
        WHERE s.status = 'active'
        GROUP BY mp.name
    `

	rows, err := r.db.QueryContext(ctx, planQuery)
	if err != nil {
		r.logger.Error("Failed to get subscriptions by plan", logger.Fields{
			"error": err.Error(),
		})
		return &stats, nil
	}
	defer rows.Close()

	stats.SubscriptionsByPlan = make(map[string]int)
	for rows.Next() {
		var planName string
		var count int
		if err := rows.Scan(&planName, &count); err == nil {
			stats.SubscriptionsByPlan[planName] = count
		}
	}

	return &stats, nil
}

func (r *subscriptionRepository) GetReactivationsCount(ctx context.Context, startDate, endDate time.Time) (int, error) {

	auditQuery := `
		SELECT COUNT(DISTINCT subscription_id)
		FROM subscription_audit 
		WHERE action = 'reactivated' 
		AND created_at BETWEEN $1 AND $2
	`

	var count int
	err := r.db.QueryRowContext(ctx, auditQuery, startDate, endDate).Scan(&count)
	if err == nil {
		r.logger.Debug("Reactivations count from audit table", logger.Fields{
			"count":      count,
			"start_date": startDate,
			"end_date":   endDate,
		})
		return count, nil
	}

	fallbackQuery := `
		SELECT COUNT(DISTINCT id)
		FROM subscriptions 
		WHERE status = 'active' 
		AND updated_at BETWEEN $1 AND $2
		AND created_at < $1
		AND updated_at > created_at + INTERVAL '1 hour'
	`

	err = r.db.QueryRowContext(ctx, fallbackQuery, startDate, endDate).Scan(&count)
	if err != nil {
		r.logger.Error("Failed to get reactivations count", logger.Fields{
			"error":      err.Error(),
			"start_date": startDate,
			"end_date":   endDate,
		})
		return 0, err
	}

	r.logger.Debug("Reactivations count from fallback method", logger.Fields{
		"count":      count,
		"start_date": startDate,
		"end_date":   endDate,
	})

	return count, nil
}

func (r *subscriptionRepository) Create(ctx context.Context, subscription *entity.Subscription) error {
	query := `
        INSERT INTO subscriptions (
            id, user_id, meal_plan_id, meal_types, delivery_days, 
            allergies, total_price, status, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    `

	_, err := r.db.ExecContext(ctx, query,
		subscription.ID, subscription.UserID, subscription.MealPlanID,
		pq.Array(subscription.MealTypes), pq.Array(subscription.DeliveryDays),
		subscription.Allergies, subscription.TotalPrice, subscription.Status,
		subscription.CreatedAt, subscription.UpdatedAt)

	if err != nil {
		r.logger.Error("Failed to create subscription", logger.Fields{
			"error":        err.Error(),
			"subscription": subscription.ID,
			"user_id":      subscription.UserID,
		})
		return err
	}

	r.LogSubscriptionAction(ctx, subscription.ID, subscription.UserID, "created", "", string(subscription.Status))

	r.logger.Info("Subscription created successfully", logger.Fields{
		"subscription": subscription.ID,
		"user_id":      subscription.UserID,
		"plan_id":      subscription.MealPlanID,
	})

	return nil
}

func (r *subscriptionRepository) GetByID(ctx context.Context, id string) (*entity.SubscriptionWithDetails, error) {
	query := `
        SELECT 
            s.id, s.user_id, s.meal_plan_id, s.meal_types, s.delivery_days,
            s.allergies, s.total_price, s.status, s.pause_start_date, s.pause_end_date,
            s.created_at, s.updated_at,
            mp.name as meal_plan_name, mp.description as meal_plan_description,
            mp.price as meal_plan_price, mp.image_url as meal_plan_image_url,
            mp.features as meal_plan_features
        FROM subscriptions s
        JOIN meal_plans mp ON s.meal_plan_id = mp.id
        WHERE s.id = $1
    `

	var sub entity.SubscriptionWithDetails
	var mealTypes pq.StringArray
	var deliveryDays pq.StringArray
	var features pq.StringArray

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&sub.ID, &sub.UserID, &sub.MealPlanID, &mealTypes, &deliveryDays,
		&sub.Allergies, &sub.TotalPrice, &sub.Status, &sub.PauseStartDate, &sub.PauseEndDate,
		&sub.CreatedAt, &sub.UpdatedAt,
		&sub.MealPlan.Name, &sub.MealPlan.Description,
		&sub.MealPlan.Price, &sub.MealPlan.ImageURL, &features,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error("Failed to get subscription by ID", logger.Fields{
			"error": err.Error(),
			"id":    id,
		})
		return nil, err
	}

	sub.MealTypes = make([]entity.MealType, len(mealTypes))
	for i, mt := range mealTypes {
		sub.MealTypes[i] = entity.MealType(mt)
	}

	sub.DeliveryDays = make([]entity.DeliveryDay, len(deliveryDays))
	for i, dd := range deliveryDays {
		sub.DeliveryDays[i] = entity.DeliveryDay(dd)
	}

	sub.MealPlan.Features = []string(features)
	sub.MealPlan.ID = sub.MealPlanID

	return &sub, nil
}

func (r *subscriptionRepository) GetByUserID(ctx context.Context, userID string) ([]entity.SubscriptionWithDetails, error) {
	query := `
        SELECT 
            s.id, s.user_id, s.meal_plan_id, s.meal_types, s.delivery_days,
            s.allergies, s.total_price, s.status, s.pause_start_date, s.pause_end_date,
            s.created_at, s.updated_at,
            mp.name as meal_plan_name, mp.description as meal_plan_description,
            mp.price as meal_plan_price, mp.image_url as meal_plan_image_url,
            mp.features as meal_plan_features
        FROM subscriptions s
        JOIN meal_plans mp ON s.meal_plan_id = mp.id
        WHERE s.user_id = $1
        ORDER BY s.created_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		r.logger.Error("Failed to get user subscriptions", logger.Fields{
			"error":   err.Error(),
			"user_id": userID,
		})
		return nil, err
	}
	defer rows.Close()

	var subscriptions []entity.SubscriptionWithDetails
	for rows.Next() {
		var sub entity.SubscriptionWithDetails
		var mealTypes pq.StringArray
		var deliveryDays pq.StringArray
		var features pq.StringArray

		err := rows.Scan(
			&sub.ID, &sub.UserID, &sub.MealPlanID, &mealTypes, &deliveryDays,
			&sub.Allergies, &sub.TotalPrice, &sub.Status, &sub.PauseStartDate, &sub.PauseEndDate,
			&sub.CreatedAt, &sub.UpdatedAt,
			&sub.MealPlan.Name, &sub.MealPlan.Description,
			&sub.MealPlan.Price, &sub.MealPlan.ImageURL, &features,
		)
		if err != nil {
			r.logger.Error("Failed to scan subscription row", logger.Fields{
				"error": err.Error(),
			})
			return nil, err
		}

		sub.MealTypes = make([]entity.MealType, len(mealTypes))
		for i, mt := range mealTypes {
			sub.MealTypes[i] = entity.MealType(mt)
		}

		sub.DeliveryDays = make([]entity.DeliveryDay, len(deliveryDays))
		for i, dd := range deliveryDays {
			sub.DeliveryDays[i] = entity.DeliveryDay(dd)
		}

		sub.MealPlan.Features = []string(features)
		sub.MealPlan.ID = sub.MealPlanID

		subscriptions = append(subscriptions, sub)
	}

	return subscriptions, nil
}

func (r *subscriptionRepository) Update(ctx context.Context, subscription *entity.Subscription) error {

	var oldStatus string
	statusQuery := `SELECT status FROM subscriptions WHERE id = $1`
	r.db.QueryRowContext(ctx, statusQuery, subscription.ID).Scan(&oldStatus)

	query := `
        UPDATE subscriptions 
        SET meal_types = $2, delivery_days = $3, allergies = $4, 
            total_price = $5, status = $6, pause_start_date = $7, 
            pause_end_date = $8, updated_at = $9
        WHERE id = $1
    `

	result, err := r.db.ExecContext(ctx, query,
		subscription.ID, pq.Array(subscription.MealTypes), pq.Array(subscription.DeliveryDays),
		subscription.Allergies, subscription.TotalPrice, subscription.Status,
		subscription.PauseStartDate, subscription.PauseEndDate, time.Now())

	if err != nil {
		r.logger.Error("Failed to update subscription", logger.Fields{
			"error":        err.Error(),
			"subscription": subscription.ID,
		})
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	action := "updated"
	if oldStatus != string(subscription.Status) {
		if oldStatus == "cancelled" && subscription.Status == entity.StatusActive {
			action = "reactivated"
		} else if subscription.Status == entity.StatusCancelled {
			action = "cancelled"
		} else if subscription.Status == entity.StatusPaused {
			action = "paused"
		} else if subscription.Status == entity.StatusActive && oldStatus == "paused" {
			action = "resumed"
		}
	}

	r.LogSubscriptionAction(ctx, subscription.ID, subscription.UserID, action, oldStatus, string(subscription.Status))

	r.logger.Info("Subscription updated successfully", logger.Fields{
		"subscription": subscription.ID,
		"old_status":   oldStatus,
		"new_status":   subscription.Status,
		"action":       action,
	})

	return nil
}

func (r *subscriptionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM subscriptions WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete subscription", logger.Fields{
			"error": err.Error(),
			"id":    id,
		})
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	r.logger.Info("Subscription deleted successfully", logger.Fields{
		"subscription": id,
	})

	return nil
}

func (r *subscriptionRepository) GetActiveSubscriptions(ctx context.Context) ([]entity.SubscriptionWithDetails, error) {
	query := `
        SELECT 
            s.id, s.user_id, s.meal_plan_id, s.meal_types, s.delivery_days,
            s.allergies, s.total_price, s.status, s.pause_start_date, s.pause_end_date,
            s.created_at, s.updated_at,
            mp.name as meal_plan_name, mp.description as meal_plan_description,
            mp.price as meal_plan_price, mp.image_url as meal_plan_image_url,
            mp.features as meal_plan_features
        FROM subscriptions s
        JOIN meal_plans mp ON s.meal_plan_id = mp.id
        WHERE s.status = 'active'
        ORDER BY s.created_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("Failed to get active subscriptions", logger.Fields{
			"error": err.Error(),
		})
		return nil, err
	}
	defer rows.Close()

	var subscriptions []entity.SubscriptionWithDetails
	for rows.Next() {
		var sub entity.SubscriptionWithDetails
		var mealTypes pq.StringArray
		var deliveryDays pq.StringArray
		var features pq.StringArray

		err := rows.Scan(
			&sub.ID, &sub.UserID, &sub.MealPlanID, &mealTypes, &deliveryDays,
			&sub.Allergies, &sub.TotalPrice, &sub.Status, &sub.PauseStartDate, &sub.PauseEndDate,
			&sub.CreatedAt, &sub.UpdatedAt,
			&sub.MealPlan.Name, &sub.MealPlan.Description,
			&sub.MealPlan.Price, &sub.MealPlan.ImageURL, &features,
		)
		if err != nil {
			return nil, err
		}

		sub.MealTypes = make([]entity.MealType, len(mealTypes))
		for i, mt := range mealTypes {
			sub.MealTypes[i] = entity.MealType(mt)
		}

		sub.DeliveryDays = make([]entity.DeliveryDay, len(deliveryDays))
		for i, dd := range deliveryDays {
			sub.DeliveryDays[i] = entity.DeliveryDay(dd)
		}

		sub.MealPlan.Features = []string(features)
		sub.MealPlan.ID = sub.MealPlanID

		subscriptions = append(subscriptions, sub)
	}

	return subscriptions, nil
}

func (r *subscriptionRepository) ExistsByUserAndPlan(ctx context.Context, userID, planID string) (bool, error) {
	query := `
        SELECT EXISTS(
            SELECT 1 FROM subscriptions 
            WHERE user_id = $1 AND meal_plan_id = $2 AND status IN ('active', 'paused')
        )
    `

	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, planID).Scan(&exists)
	if err != nil {
		r.logger.Error("Failed to check subscription existence", logger.Fields{
			"error":   err.Error(),
			"user_id": userID,
			"plan_id": planID,
		})
		return false, err
	}

	return exists, nil
}

func (r *subscriptionRepository) GetExpiredSubscriptions(ctx context.Context) ([]entity.Subscription, error) {
	query := `
        SELECT id, user_id, meal_plan_id, meal_types, delivery_days,
               allergies, total_price, status, pause_start_date, pause_end_date,
               created_at, updated_at
        FROM subscriptions
        WHERE status = 'paused' AND pause_end_date < NOW()
    `

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("Failed to get expired subscriptions", logger.Fields{
			"error": err.Error(),
		})
		return nil, err
	}
	defer rows.Close()

	var subscriptions []entity.Subscription
	for rows.Next() {
		var sub entity.Subscription
		var mealTypes pq.StringArray
		var deliveryDays pq.StringArray

		err := rows.Scan(
			&sub.ID, &sub.UserID, &sub.MealPlanID, &mealTypes, &deliveryDays,
			&sub.Allergies, &sub.TotalPrice, &sub.Status, &sub.PauseStartDate, &sub.PauseEndDate,
			&sub.CreatedAt, &sub.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		sub.MealTypes = make([]entity.MealType, len(mealTypes))
		for i, mt := range mealTypes {
			sub.MealTypes[i] = entity.MealType(mt)
		}

		sub.DeliveryDays = make([]entity.DeliveryDay, len(deliveryDays))
		for i, dd := range deliveryDays {
			sub.DeliveryDays[i] = entity.DeliveryDay(dd)
		}

		subscriptions = append(subscriptions, sub)
	}

	return subscriptions, nil
}

func (r *subscriptionRepository) BulkUpdateStatus(ctx context.Context, ids []string, status entity.SubscriptionStatus) error {
	if len(ids) == 0 {
		return nil
	}

	query := `
        UPDATE subscriptions 
        SET status = $1, updated_at = $2
        WHERE id = ANY($3)
    `

	_, err := r.db.ExecContext(ctx, query, status, time.Now(), pq.Array(ids))
	if err != nil {
		r.logger.Error("Failed to bulk update subscription status", logger.Fields{
			"error":  err.Error(),
			"ids":    ids,
			"status": status,
		})
		return err
	}

	r.logger.Info("Bulk updated subscription status", logger.Fields{
		"count":  len(ids),
		"status": status,
	})

	return nil
}
