package service

import (
	"context"
	"fmt"
	"time"

	"sea-catering-backend/internal/api/meal_plans/repository"
	"sea-catering-backend/internal/api/subscriptions"
	subscriptionRepo "sea-catering-backend/internal/api/subscriptions/repository"
	"sea-catering-backend/internal/entity"
	"sea-catering-backend/pkg/logger"
	"sea-catering-backend/pkg/utils"
)

type SubscriptionService interface {
	CreateSubscription(ctx context.Context, req subscriptions.CreateSubscriptionRequest) (*entity.SubscriptionWithDetails, error)
	GetUserSubscriptions(ctx context.Context, userID string) ([]entity.SubscriptionWithDetails, error)
	GetSubscriptionByID(ctx context.Context, subscriptionID string) (*entity.SubscriptionWithDetails, error)
	PauseSubscription(ctx context.Context, subscriptionID, userID string, startDate, endDate time.Time) error
	ResumeSubscription(ctx context.Context, subscriptionID, userID string) error
	CancelSubscription(ctx context.Context, subscriptionID, userID string) error
	ReactivateSubscription(ctx context.Context, subscriptionID, userID string) (*entity.SubscriptionWithDetails, error)
	UpdateSubscription(ctx context.Context, subscriptionID, userID string, req subscriptions.CreateSubscriptionRequest) (*entity.SubscriptionWithDetails, error)
	GetSubscriptionStats(ctx context.Context, startDate, endDate time.Time) (*subscriptions.SubscriptionStatsResponse, error)
	ProcessExpiredPauses(ctx context.Context) error
}

type subscriptionService struct {
	subscriptionRepo subscriptionRepo.SubscriptionRepository
	mealPlanRepo     repository.MealPlanRepository
	utils            utils.Interface
	logger           *logger.Logger
}

func NewSubscriptionService(
	subscriptionRepo subscriptionRepo.SubscriptionRepository,
	mealPlanRepo repository.MealPlanRepository,
	utils utils.Interface,
	logger *logger.Logger,
) SubscriptionService {
	return &subscriptionService{
		subscriptionRepo: subscriptionRepo,
		mealPlanRepo:     mealPlanRepo,
		utils:            utils,
		logger:           logger,
	}
}

func (s *subscriptionService) CreateSubscription(ctx context.Context, req subscriptions.CreateSubscriptionRequest) (*entity.SubscriptionWithDetails, error) {

	if err := s.utils.ValidateMealTypes(convertMealTypesToStrings(req.MealTypes)); err != nil {
		return nil, subscriptions.ErrInvalidMealTypes
	}

	if err := s.utils.ValidateDeliveryDays(convertDeliveryDaysToStrings(req.DeliveryDays)); err != nil {
		return nil, subscriptions.ErrInvalidDeliveryDays
	}

	mealPlan, err := s.mealPlanRepo.GetByID(ctx, req.MealPlanID)
	if err != nil {
		s.logger.Error("Failed to get meal plan", logger.Fields{
			"error":   err.Error(),
			"plan_id": req.MealPlanID,
		})
		return nil, subscriptions.ErrInvalidMealPlan
	}

	if mealPlan == nil {
		return nil, subscriptions.ErrInvalidMealPlan
	}

	totalPrice := s.utils.CalculateSubscriptionPrice(
		mealPlan.Price,
		convertMealTypesToStrings(req.MealTypes),
		convertDeliveryDaysToStrings(req.DeliveryDays),
	)

	subscriptionID := s.utils.GenerateULID()
	userID := ctx.Value("user_id").(string)

	subscription := &entity.Subscription{
		ID:           subscriptionID,
		UserID:       userID,
		MealPlanID:   req.MealPlanID,
		MealTypes:    req.MealTypes,
		DeliveryDays: req.DeliveryDays,
		Allergies:    req.Allergies,
		TotalPrice:   totalPrice,
		Status:       entity.StatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.subscriptionRepo.Create(ctx, subscription); err != nil {
		s.logger.Error("Failed to create subscription", logger.Fields{
			"error":        err.Error(),
			"subscription": subscriptionID,
		})
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	subscriptionDetails, err := s.subscriptionRepo.GetByID(ctx, subscriptionID)
	if err != nil {
		s.logger.Error("Failed to get subscription details", logger.Fields{
			"error":        err.Error(),
			"subscription": subscriptionID,
		})
		return nil, fmt.Errorf("failed to get subscription details: %w", err)
	}

	s.logger.Info("Subscription created successfully", logger.Fields{
		"subscription": subscriptionID,
		"user_id":      userID,
		"plan":         mealPlan.Name,
		"total_price":  totalPrice,
		"status":       "active",
	})

	return subscriptionDetails, nil
}

func (s *subscriptionService) ReactivateSubscription(ctx context.Context, subscriptionID, userID string) (*entity.SubscriptionWithDetails, error) {
	s.logger.Info("Starting subscription reactivation", logger.Fields{
		"subscription_id": subscriptionID,
		"user_id":         userID,
	})

	subscription, err := s.subscriptionRepo.GetSubscriptionByIDForReactivation(ctx, subscriptionID)
	if err != nil {
		s.logger.Error("Failed to get subscription for reactivation", logger.Fields{
			"error":           err.Error(),
			"subscription_id": subscriptionID,
		})
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	if subscription.UserID != userID {
		s.logger.Warn("Unauthorized reactivation attempt", logger.Fields{
			"subscription_id":    subscriptionID,
			"requesting_user_id": userID,
			"actual_user_id":     subscription.UserID,
		})
		return nil, subscriptions.ErrUnauthorizedAccess
	}

	if subscription.Status != entity.StatusCancelled {
		s.logger.Warn("Cannot reactivate subscription with status", logger.Fields{
			"subscription_id": subscriptionID,
			"current_status":  subscription.Status,
		})
		return nil, fmt.Errorf("only cancelled subscriptions can be reactivated, current status: %s", subscription.Status)
	}

	oldStatus := subscription.Status
	subscription.Status = entity.StatusActive
	subscription.PauseStartDate = nil
	subscription.PauseEndDate = nil
	subscription.UpdatedAt = time.Now()

	if err := s.subscriptionRepo.Update(ctx, subscription); err != nil {
		s.logger.Error("Failed to update subscription for reactivation", logger.Fields{
			"error":           err.Error(),
			"subscription_id": subscriptionID,
		})
		return nil, fmt.Errorf("failed to reactivate subscription: %w", err)
	}

	subscriptionWithDetails, err := s.subscriptionRepo.GetByID(ctx, subscriptionID)
	if err != nil {
		s.logger.Error("Failed to get updated subscription details", logger.Fields{
			"error":           err.Error(),
			"subscription_id": subscriptionID,
		})

		mealPlan, _ := s.mealPlanRepo.GetByID(ctx, subscription.MealPlanID)
		if mealPlan != nil {
			subscriptionWithDetails = &entity.SubscriptionWithDetails{
				Subscription: *subscription,
				MealPlan:     *mealPlan,
			}
		}
	}

	s.logger.Info("Subscription reactivated successfully", logger.Fields{
		"subscription_id": subscriptionID,
		"user_id":         userID,
		"old_status":      oldStatus,
		"new_status":      subscription.Status,
	})

	return subscriptionWithDetails, nil
}

func (s *subscriptionService) GetUserSubscriptions(ctx context.Context, userID string) ([]entity.SubscriptionWithDetails, error) {
	subscriptions, err := s.subscriptionRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user subscriptions", logger.Fields{
			"error":   err.Error(),
			"user_id": userID,
		})
		return nil, fmt.Errorf("failed to get user subscriptions: %w", err)
	}

	return subscriptions, nil
}

func (s *subscriptionService) GetSubscriptionByID(ctx context.Context, subscriptionID string) (*entity.SubscriptionWithDetails, error) {
	subscription, err := s.subscriptionRepo.GetByID(ctx, subscriptionID)
	if err != nil {
		s.logger.Error("Failed to get subscription", logger.Fields{
			"error":        err.Error(),
			"subscription": subscriptionID,
		})
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	if subscription == nil {
		return nil, subscriptions.ErrSubscriptionNotFound
	}

	return subscription, nil
}

func (s *subscriptionService) PauseSubscription(ctx context.Context, subscriptionID, userID string, startDate, endDate time.Time) error {

	if startDate.After(endDate) || startDate.Before(time.Now()) {
		return subscriptions.ErrInvalidPauseDates
	}

	subscription, err := s.subscriptionRepo.GetByID(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if subscription == nil {
		return subscriptions.ErrSubscriptionNotFound
	}

	if subscription.UserID != userID {
		return subscriptions.ErrUnauthorizedAccess
	}

	if subscription.Status != entity.StatusActive {
		return subscriptions.ErrSubscriptionCancelled
	}

	subscription.Status = entity.StatusPaused
	subscription.PauseStartDate = &startDate
	subscription.PauseEndDate = &endDate

	if err := s.subscriptionRepo.Update(ctx, &subscription.Subscription); err != nil {
		s.logger.Error("Failed to pause subscription", logger.Fields{
			"error":        err.Error(),
			"subscription": subscriptionID,
		})
		return fmt.Errorf("failed to pause subscription: %w", err)
	}

	s.logger.Info("Subscription paused successfully", logger.Fields{
		"subscription": subscriptionID,
		"user_id":      userID,
		"start_date":   startDate,
		"end_date":     endDate,
	})

	return nil
}

func (s *subscriptionService) ResumeSubscription(ctx context.Context, subscriptionID, userID string) error {
	subscription, err := s.subscriptionRepo.GetByID(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if subscription == nil {
		return subscriptions.ErrSubscriptionNotFound
	}

	if subscription.UserID != userID {
		return subscriptions.ErrUnauthorizedAccess
	}

	if subscription.Status != entity.StatusPaused {
		return fmt.Errorf("subscription is not paused")
	}

	subscription.Status = entity.StatusActive
	subscription.PauseStartDate = nil
	subscription.PauseEndDate = nil

	if err := s.subscriptionRepo.Update(ctx, &subscription.Subscription); err != nil {
		s.logger.Error("Failed to resume subscription", logger.Fields{
			"error":        err.Error(),
			"subscription": subscriptionID,
		})
		return fmt.Errorf("failed to resume subscription: %w", err)
	}

	s.logger.Info("Subscription resumed successfully", logger.Fields{
		"subscription": subscriptionID,
		"user_id":      userID,
	})

	return nil
}

func (s *subscriptionService) CancelSubscription(ctx context.Context, subscriptionID, userID string) error {
	subscription, err := s.subscriptionRepo.GetByID(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if subscription == nil {
		return subscriptions.ErrSubscriptionNotFound
	}

	if subscription.UserID != userID {
		return subscriptions.ErrUnauthorizedAccess
	}

	if subscription.Status == entity.StatusCancelled {
		return fmt.Errorf("subscription is already cancelled")
	}

	subscription.Status = entity.StatusCancelled
	subscription.PauseStartDate = nil
	subscription.PauseEndDate = nil

	if err := s.subscriptionRepo.Update(ctx, &subscription.Subscription); err != nil {
		s.logger.Error("Failed to cancel subscription", logger.Fields{
			"error":        err.Error(),
			"subscription": subscriptionID,
		})
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	s.logger.Info("Subscription cancelled successfully", logger.Fields{
		"subscription": subscriptionID,
		"user_id":      userID,
	})

	return nil
}

func (s *subscriptionService) UpdateSubscription(ctx context.Context, subscriptionID, userID string, req subscriptions.CreateSubscriptionRequest) (*entity.SubscriptionWithDetails, error) {

	if err := s.utils.ValidateMealTypes(convertMealTypesToStrings(req.MealTypes)); err != nil {
		return nil, subscriptions.ErrInvalidMealTypes
	}

	if err := s.utils.ValidateDeliveryDays(convertDeliveryDaysToStrings(req.DeliveryDays)); err != nil {
		return nil, subscriptions.ErrInvalidDeliveryDays
	}

	subscription, err := s.subscriptionRepo.GetByID(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	if subscription == nil {
		return nil, subscriptions.ErrSubscriptionNotFound
	}

	if subscription.UserID != userID {
		return nil, subscriptions.ErrUnauthorizedAccess
	}

	if subscription.Status == entity.StatusCancelled {
		return nil, fmt.Errorf("cannot update cancelled subscription")
	}

	mealPlan, err := s.mealPlanRepo.GetByID(ctx, req.MealPlanID)
	if err != nil {
		return nil, subscriptions.ErrInvalidMealPlan
	}

	if mealPlan == nil {
		return nil, subscriptions.ErrInvalidMealPlan
	}

	totalPrice := s.utils.CalculateSubscriptionPrice(
		mealPlan.Price,
		convertMealTypesToStrings(req.MealTypes),
		convertDeliveryDaysToStrings(req.DeliveryDays),
	)

	subscription.MealPlanID = req.MealPlanID
	subscription.MealTypes = req.MealTypes
	subscription.DeliveryDays = req.DeliveryDays
	subscription.Allergies = req.Allergies
	subscription.TotalPrice = totalPrice

	if err := s.subscriptionRepo.Update(ctx, &subscription.Subscription); err != nil {
		s.logger.Error("Failed to update subscription", logger.Fields{
			"error":        err.Error(),
			"subscription": subscriptionID,
		})
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	updatedSubscription, err := s.subscriptionRepo.GetByID(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated subscription: %w", err)
	}

	s.logger.Info("Subscription updated successfully", logger.Fields{
		"subscription": subscriptionID,
		"user_id":      userID,
		"new_price":    totalPrice,
	})

	return updatedSubscription, nil
}

func (s *subscriptionService) GetSubscriptionStats(ctx context.Context, startDate, endDate time.Time) (*subscriptions.SubscriptionStatsResponse, error) {
	stats, err := s.subscriptionRepo.GetSubscriptionStats(ctx, startDate, endDate)
	if err != nil {
		s.logger.Error("Failed to get subscription stats", logger.Fields{
			"error":      err.Error(),
			"start_date": startDate,
			"end_date":   endDate,
		})
		return nil, fmt.Errorf("failed to get subscription stats: %w", err)
	}

	reactivations, err := s.subscriptionRepo.GetReactivationsCount(ctx, startDate, endDate)
	if err != nil {
		s.logger.Warn("Failed to get reactivations count", logger.Fields{
			"error": err.Error(),
		})
		reactivations = 0
	}

	response := &subscriptions.SubscriptionStatsResponse{
		TotalSubscriptions:     stats.TotalSubscriptions,
		ActiveSubscriptions:    stats.ActiveSubscriptions,
		PausedSubscriptions:    stats.PausedSubscriptions,
		CancelledSubscriptions: stats.CancelledSubscriptions,
		MonthlyRevenue:         stats.MonthlyRevenue,
		NewSubscriptions:       stats.NewSubscriptions,
		Reactivations:          reactivations,
		SubscriptionsByPlan:    stats.SubscriptionsByPlan,
	}

	return response, nil
}

func (s *subscriptionService) ProcessExpiredPauses(ctx context.Context) error {

	expiredSubscriptions, err := s.subscriptionRepo.GetExpiredSubscriptions(ctx)
	if err != nil {
		s.logger.Error("Failed to get expired subscriptions", logger.Fields{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to get expired subscriptions: %w", err)
	}

	if len(expiredSubscriptions) == 0 {
		return nil
	}

	var ids []string
	for _, sub := range expiredSubscriptions {
		ids = append(ids, sub.ID)
	}

	if err := s.subscriptionRepo.BulkUpdateStatus(ctx, ids, entity.StatusActive); err != nil {
		s.logger.Error("Failed to bulk resume expired subscriptions", logger.Fields{
			"error": err.Error(),
			"count": len(ids),
		})
		return fmt.Errorf("failed to bulk resume subscriptions: %w", err)
	}

	s.logger.Info("Processed expired paused subscriptions", logger.Fields{
		"count": len(ids),
	})

	return nil
}

func convertMealTypesToStrings(mealTypes []entity.MealType) []string {
	result := make([]string, len(mealTypes))
	for i, mt := range mealTypes {
		result[i] = string(mt)
	}
	return result
}

func convertDeliveryDaysToStrings(deliveryDays []entity.DeliveryDay) []string {
	result := make([]string, len(deliveryDays))
	for i, dd := range deliveryDays {
		result[i] = string(dd)
	}
	return result
}
