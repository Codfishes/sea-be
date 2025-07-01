package service

import (
	"context"
	"strings"
	"time"

	"sea-catering-backend/internal/api/meal_plans"
	"sea-catering-backend/internal/api/meal_plans/repository"
	"sea-catering-backend/internal/entity"
	"sea-catering-backend/pkg/logger"
	"sea-catering-backend/pkg/utils"
)

type MealPlanService interface {
	CreateMealPlan(ctx context.Context, req meal_plans.CreateMealPlanRequest) (*meal_plans.MealPlanResponse, error)
	GetMealPlanByID(ctx context.Context, id string) (*meal_plans.MealPlanResponse, error)
	UpdateMealPlan(ctx context.Context, id string, req meal_plans.UpdateMealPlanRequest) (*meal_plans.MealPlanResponse, error)
	DeleteMealPlan(ctx context.Context, id string) error

	GetAllMealPlans(ctx context.Context, params meal_plans.MealPlanListRequest) (*meal_plans.MealPlanListResponse, error)
	GetActiveMealPlans(ctx context.Context) ([]meal_plans.MealPlanResponse, error)
	SearchMealPlans(ctx context.Context, query string, limit int) ([]meal_plans.MealPlanResponse, error)

	ActivateMealPlan(ctx context.Context, id string) error
	DeactivateMealPlan(ctx context.Context, id string) error
	BulkUpdateActiveStatus(ctx context.Context, ids []string, isActive bool) error

	GetMealPlanStats(ctx context.Context) (*meal_plans.MealPlanStatsResponse, error)
	GetPopularityRanking(ctx context.Context) ([]meal_plans.MealPlanResponse, error)

	ValidateMealPlanAccess(ctx context.Context, id string) error
	CheckMealPlanAvailability(ctx context.Context, id string) error
}

type mealPlanService struct {
	repo         repository.MealPlanRepository
	utilsService utils.Interface
	logger       *logger.Logger
}

func NewMealPlanService(
	repo repository.MealPlanRepository,
	utilsService utils.Interface,
	logger *logger.Logger,
) MealPlanService {
	return &mealPlanService{
		repo:         repo,
		utilsService: utilsService,
		logger:       logger,
	}
}

func (s *mealPlanService) CreateMealPlan(ctx context.Context, req meal_plans.CreateMealPlanRequest) (*meal_plans.MealPlanResponse, error) {

	if err := s.validateCreateRequest(ctx, req); err != nil {
		return nil, err
	}

	exists, err := s.repo.ExistsByName(ctx, req.Name, "")
	if err != nil {
		s.logger.Error("Failed to check meal plan name existence", logger.Fields{
			"name":  req.Name,
			"error": err.Error(),
		})
		return nil, err
	}

	if exists {
		return nil, meal_plans.ErrMealPlanNameTaken
	}

	mealPlan := &entity.MealPlan{
		ID:          s.utilsService.GenerateULID(),
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Price:       req.Price,
		ImageURL:    req.ImageURL,
		Features:    s.cleanFeatures(req.Features),
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, mealPlan); err != nil {
		s.logger.Error("Failed to create meal plan", logger.Fields{
			"name":  req.Name,
			"error": err.Error(),
		})
		return nil, err
	}

	s.logger.Info("Meal plan created successfully", logger.Fields{
		"id":   mealPlan.ID,
		"name": mealPlan.Name,
	})

	return s.entityToResponse(mealPlan), nil
}

func (s *mealPlanService) GetMealPlanByID(ctx context.Context, id string) (*meal_plans.MealPlanResponse, error) {
	if id == "" {
		return nil, meal_plans.ErrMealPlanNotFound
	}

	mealPlan, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get meal plan by ID", logger.Fields{
			"id":    id,
			"error": err.Error(),
		})
		return nil, err
	}

	return s.entityToResponse(mealPlan), nil
}

func (s *mealPlanService) UpdateMealPlan(ctx context.Context, id string, req meal_plans.UpdateMealPlanRequest) (*meal_plans.MealPlanResponse, error) {

	existingPlan, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.validateUpdateRequest(ctx, req, id); err != nil {
		return nil, err
	}

	updatedPlan := s.applyUpdates(existingPlan, req)
	updatedPlan.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, updatedPlan); err != nil {
		s.logger.Error("Failed to update meal plan", logger.Fields{
			"id":    id,
			"error": err.Error(),
		})
		return nil, err
	}

	s.logger.Info("Meal plan updated successfully", logger.Fields{
		"id":   id,
		"name": updatedPlan.Name,
	})

	return s.entityToResponse(updatedPlan), nil
}

func (s *mealPlanService) DeleteMealPlan(ctx context.Context, id string) error {

	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete meal plan", logger.Fields{
			"id":    id,
			"error": err.Error(),
		})
		return err
	}

	s.logger.Info("Meal plan deleted successfully", logger.Fields{
		"id": id,
	})

	return nil
}

func (s *mealPlanService) GetAllMealPlans(ctx context.Context, params meal_plans.MealPlanListRequest) (*meal_plans.MealPlanListResponse, error) {

	if params.Page < 0 || params.Limit < 0 || params.Limit > 100 {
		return nil, meal_plans.ErrInvalidPaginationParams
	}

	mealPlans, meta, err := s.repo.List(ctx, params)
	if err != nil {
		s.logger.Error("Failed to get meal plans list", logger.Fields{
			"params": params,
			"error":  err.Error(),
		})
		return nil, err
	}

	responses := make([]meal_plans.MealPlanResponse, len(mealPlans))
	for i, plan := range mealPlans {
		responses[i] = *s.entityToResponse(&plan)
	}

	return &meal_plans.MealPlanListResponse{
		MealPlans: responses,
		Meta:      meta,
	}, nil
}

func (s *mealPlanService) GetActiveMealPlans(ctx context.Context) ([]meal_plans.MealPlanResponse, error) {
	mealPlans, err := s.repo.GetActive(ctx)
	if err != nil {
		s.logger.Error("Failed to get active meal plans", logger.Fields{
			"error": err.Error(),
		})
		return nil, err
	}

	responses := make([]meal_plans.MealPlanResponse, len(mealPlans))
	for i, plan := range mealPlans {
		responses[i] = *s.entityToResponse(&plan)
	}

	return responses, nil
}

func (s *mealPlanService) SearchMealPlans(ctx context.Context, query string, limit int) ([]meal_plans.MealPlanResponse, error) {
	if query == "" {
		return []meal_plans.MealPlanResponse{}, nil
	}

	searchQuery := strings.TrimSpace(query)
	if len(searchQuery) < 2 {
		return []meal_plans.MealPlanResponse{}, nil
	}

	mealPlans, err := s.repo.Search(ctx, searchQuery, limit)
	if err != nil {
		s.logger.Error("Failed to search meal plans", logger.Fields{
			"query": searchQuery,
			"error": err.Error(),
		})
		return nil, err
	}

	responses := make([]meal_plans.MealPlanResponse, len(mealPlans))
	for i, plan := range mealPlans {
		responses[i] = *s.entityToResponse(&plan)
	}

	return responses, nil
}

func (s *mealPlanService) ActivateMealPlan(ctx context.Context, id string) error {
	mealPlan, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if mealPlan.IsActive {
		return nil
	}

	mealPlan.IsActive = true
	mealPlan.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, mealPlan); err != nil {
		s.logger.Error("Failed to activate meal plan", logger.Fields{
			"id":    id,
			"error": err.Error(),
		})
		return err
	}

	s.logger.Info("Meal plan activated", logger.Fields{
		"id":   id,
		"name": mealPlan.Name,
	})

	return nil
}

func (s *mealPlanService) DeactivateMealPlan(ctx context.Context, id string) error {
	mealPlan, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if !mealPlan.IsActive {
		return nil
	}

	mealPlan.IsActive = false
	mealPlan.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, mealPlan); err != nil {
		s.logger.Error("Failed to deactivate meal plan", logger.Fields{
			"id":    id,
			"error": err.Error(),
		})
		return err
	}

	s.logger.Info("Meal plan deactivated", logger.Fields{
		"id":   id,
		"name": mealPlan.Name,
	})

	return nil
}

func (s *mealPlanService) BulkUpdateActiveStatus(ctx context.Context, ids []string, isActive bool) error {
	if len(ids) == 0 {
		return nil
	}

	for _, id := range ids {
		exists, err := s.repo.ExistsByID(ctx, id)
		if err != nil {
			return err
		}
		if !exists {
			return meal_plans.ErrMealPlanNotFound
		}
	}

	if err := s.repo.UpdateActiveStatus(ctx, ids, isActive); err != nil {
		s.logger.Error("Failed to bulk update meal plan status", logger.Fields{
			"ids":       ids,
			"is_active": isActive,
			"error":     err.Error(),
		})
		return err
	}

	action := "activated"
	if !isActive {
		action = "deactivated"
	}

	s.logger.Info("Meal plans bulk updated", logger.Fields{
		"count":  len(ids),
		"action": action,
	})

	return nil
}

func (s *mealPlanService) GetMealPlanStats(ctx context.Context) (*meal_plans.MealPlanStatsResponse, error) {
	stats, err := s.repo.GetStats(ctx)
	if err != nil {
		s.logger.Error("Failed to get meal plan stats", logger.Fields{
			"error": err.Error(),
		})
		return nil, err
	}

	return stats, nil
}

func (s *mealPlanService) GetPopularityRanking(ctx context.Context) ([]meal_plans.MealPlanResponse, error) {
	mealPlans, err := s.repo.GetPopularityRanking(ctx)
	if err != nil {
		s.logger.Error("Failed to get popularity ranking", logger.Fields{
			"error": err.Error(),
		})
		return nil, err
	}

	responses := make([]meal_plans.MealPlanResponse, len(mealPlans))
	for i, plan := range mealPlans {
		responses[i] = *s.entityToResponse(&plan)
	}

	return responses, nil
}

func (s *mealPlanService) ValidateMealPlanAccess(ctx context.Context, id string) error {
	exists, err := s.repo.ExistsByID(ctx, id)
	if err != nil {
		return err
	}

	if !exists {
		return meal_plans.ErrMealPlanNotFound
	}

	return nil
}

func (s *mealPlanService) CheckMealPlanAvailability(ctx context.Context, id string) error {
	isActive, err := s.repo.IsActive(ctx, id)
	if err != nil {
		return err
	}

	if !isActive {
		return meal_plans.ErrMealPlanInactive
	}

	return nil
}

func (s *mealPlanService) validateCreateRequest(ctx context.Context, req meal_plans.CreateMealPlanRequest) error {
	if req.Price <= 0 {
		return meal_plans.ErrInvalidPrice
	}

	if len(req.Features) == 0 {
		return meal_plans.ErrInvalidFeatures
	}

	if req.ImageURL != "" && !s.utilsService.IsValidURL(req.ImageURL) {
		return meal_plans.ErrInvalidImageURL
	}

	return nil
}

func (s *mealPlanService) validateUpdateRequest(ctx context.Context, req meal_plans.UpdateMealPlanRequest, excludeID string) error {
	if req.Price != nil && *req.Price <= 0 {
		return meal_plans.ErrInvalidPrice
	}

	if req.Features != nil && len(*req.Features) == 0 {
		return meal_plans.ErrInvalidFeatures
	}

	if req.ImageURL != nil && *req.ImageURL != "" && !s.utilsService.IsValidURL(*req.ImageURL) {
		return meal_plans.ErrInvalidImageURL
	}

	if req.Name != nil {
		exists, err := s.repo.ExistsByName(ctx, *req.Name, excludeID)
		if err != nil {
			return err
		}
		if exists {
			return meal_plans.ErrMealPlanNameTaken
		}
	}

	return nil
}

func (s *mealPlanService) applyUpdates(existing *entity.MealPlan, req meal_plans.UpdateMealPlanRequest) *entity.MealPlan {
	updated := *existing

	if req.Name != nil {
		updated.Name = strings.TrimSpace(*req.Name)
	}

	if req.Description != nil {
		updated.Description = strings.TrimSpace(*req.Description)
	}

	if req.Price != nil {
		updated.Price = *req.Price
	}

	if req.ImageURL != nil {
		updated.ImageURL = *req.ImageURL
	}

	if req.Features != nil {
		updated.Features = s.cleanFeatures(*req.Features)
	}

	if req.IsActive != nil {
		updated.IsActive = *req.IsActive
	}

	return &updated
}

func (s *mealPlanService) cleanFeatures(features []string) []string {
	var cleaned []string
	seen := make(map[string]bool)

	for _, feature := range features {
		trimmed := strings.TrimSpace(feature)
		if trimmed != "" && !seen[trimmed] {
			cleaned = append(cleaned, trimmed)
			seen[trimmed] = true
		}
	}

	return cleaned
}

func (s *mealPlanService) entityToResponse(plan *entity.MealPlan) *meal_plans.MealPlanResponse {
	return &meal_plans.MealPlanResponse{
		ID:          plan.ID,
		Name:        plan.Name,
		Description: plan.Description,
		Price:       plan.Price,
		ImageURL:    plan.ImageURL,
		Features:    plan.Features,
		IsActive:    plan.IsActive,
		CreatedAt:   plan.CreatedAt,
		UpdatedAt:   plan.UpdatedAt,
	}
}
