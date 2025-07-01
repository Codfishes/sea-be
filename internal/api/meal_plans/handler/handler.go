package handler

import (
	"fmt"
	"sea-catering-backend/pkg/logger"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"sea-catering-backend/internal/api/meal_plans"
	"sea-catering-backend/internal/api/meal_plans/service"
	"sea-catering-backend/internal/middleware"
	"sea-catering-backend/pkg/context"
	"sea-catering-backend/pkg/handlerutil"
	"sea-catering-backend/pkg/response"
)

type MealPlanHandler struct {
	mealPlanService service.MealPlanService
	validator       *validator.Validate
	middleware      middleware.Interface
	logger          *logger.Logger
}

func NewMealPlanHandler(
	mealPlanService service.MealPlanService,
	validator *validator.Validate,
	middleware middleware.Interface,
	logger *logger.Logger,
) *MealPlanHandler {
	return &MealPlanHandler{
		mealPlanService: mealPlanService,
		validator:       validator,
		middleware:      middleware,
		logger:          logger,
	}
}

func (h *MealPlanHandler) RegisterRoutes(router fiber.Router) {
	plans := router.Group("/meal-plans")

	plans.Get("/", h.GetMealPlans)
	plans.Get("/active", h.GetActiveMealPlans)
	plans.Get("/search", h.SearchMealPlans)
	plans.Get("/popular", h.GetPopularityRanking)
	plans.Get("/:id", h.GetMealPlan)

	admin := plans.Group("/admin")
	admin.Use(h.middleware.AdminMiddleware())
	admin.Post("/", h.CreateMealPlan)
	admin.Put("/:id", h.UpdateMealPlan)
	admin.Delete("/:id", h.DeleteMealPlan)
	admin.Patch("/:id/activate", h.ActivateMealPlan)
	admin.Patch("/:id/deactivate", h.DeactivateMealPlan)
	admin.Patch("/bulk-status", h.BulkUpdateStatus)
	admin.Get("/stats", h.GetMealPlanStats)
}

func (h *MealPlanHandler) GetMealPlans(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.middleware.GetRequestID(c)

	var params meal_plans.MealPlanListRequest
	if err := c.QueryParser(&params); err != nil {
		return errHandler.HandleBadRequest(c, requestID, "Invalid query parameters")
	}

	if err := h.validator.Struct(params); err != nil {
		return errHandler.HandleValidationError(c, requestID, err, c.Path())
	}

	result, err := h.mealPlanService.GetAllMealPlans(ctx, params)
	if err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "get_meal_plans")
	}

	return response.Success(c, result)
}

func (h *MealPlanHandler) GetActiveMealPlans(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.middleware.GetRequestID(c)

	mealPlans, err := h.mealPlanService.GetActiveMealPlans(ctx)
	if err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "get_active_meal_plans")
	}

	return response.Success(c, mealPlans)
}

func (h *MealPlanHandler) SearchMealPlans(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.middleware.GetRequestID(c)

	query := c.Query("q", "")
	limitStr := c.Query("limit", "10")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}

	if query == "" {
		return response.Success(c, []meal_plans.MealPlanResponse{})
	}

	mealPlans, err := h.mealPlanService.SearchMealPlans(ctx, query, limit)
	if err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "search_meal_plans")
	}

	return response.Success(c, mealPlans)
}

func (h *MealPlanHandler) GetPopularityRanking(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.middleware.GetRequestID(c)

	mealPlans, err := h.mealPlanService.GetPopularityRanking(ctx)
	if err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "get_popularity_ranking")
	}

	return response.Success(c, mealPlans)
}

func (h *MealPlanHandler) GetMealPlan(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.middleware.GetRequestID(c)

	mealPlanID := c.Params("id")
	if mealPlanID == "" {
		return errHandler.HandleBadRequest(c, requestID, "Meal plan ID is required")
	}

	mealPlan, err := h.mealPlanService.GetMealPlanByID(ctx, mealPlanID)
	if err != nil {
		if err == meal_plans.ErrMealPlanNotFound {
			return errHandler.HandleNotFound(c, requestID, "Meal plan")
		}
		return errHandler.Handle(c, requestID, err, c.Path(), "get_meal_plan")
	}

	return response.Success(c, mealPlan)
}

func (h *MealPlanHandler) CreateMealPlan(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 15*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.middleware.GetRequestID(c)

	var req meal_plans.CreateMealPlanRequest
	if err := c.BodyParser(&req); err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "parse_request_body")
	}

	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(c, requestID, err, c.Path())
	}

	mealPlan, err := h.mealPlanService.CreateMealPlan(ctx, req)
	if err != nil {
		return h.handleMealPlanError(c, errHandler, requestID, err, c.Path(), "create_meal_plan")
	}

	return response.Created(c, mealPlan, "Meal plan created successfully")
}

func (h *MealPlanHandler) UpdateMealPlan(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 15*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.middleware.GetRequestID(c)

	mealPlanID := c.Params("id")
	if mealPlanID == "" {
		return errHandler.HandleBadRequest(c, requestID, "Meal plan ID is required")
	}

	var req meal_plans.UpdateMealPlanRequest
	if err := c.BodyParser(&req); err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "parse_request_body")
	}

	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(c, requestID, err, c.Path())
	}

	mealPlan, err := h.mealPlanService.UpdateMealPlan(ctx, mealPlanID, req)
	if err != nil {
		return h.handleMealPlanError(c, errHandler, requestID, err, c.Path(), "update_meal_plan")
	}

	return response.Updated(c, mealPlan, "Meal plan updated successfully")
}

func (h *MealPlanHandler) DeleteMealPlan(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 15*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.middleware.GetRequestID(c)

	mealPlanID := c.Params("id")
	if mealPlanID == "" {
		return errHandler.HandleBadRequest(c, requestID, "Meal plan ID is required")
	}

	err := h.mealPlanService.DeleteMealPlan(ctx, mealPlanID)
	if err != nil {
		return h.handleMealPlanError(c, errHandler, requestID, err, c.Path(), "delete_meal_plan")
	}

	return response.Deleted(c, "Meal plan deleted successfully")
}

func (h *MealPlanHandler) ActivateMealPlan(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.middleware.GetRequestID(c)

	mealPlanID := c.Params("id")
	if mealPlanID == "" {
		return errHandler.HandleBadRequest(c, requestID, "Meal plan ID is required")
	}

	err := h.mealPlanService.ActivateMealPlan(ctx, mealPlanID)
	if err != nil {
		return h.handleMealPlanError(c, errHandler, requestID, err, c.Path(), "activate_meal_plan")
	}

	return response.Success(c, nil, "Meal plan activated successfully")
}

func (h *MealPlanHandler) DeactivateMealPlan(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.middleware.GetRequestID(c)

	mealPlanID := c.Params("id")
	if mealPlanID == "" {
		return errHandler.HandleBadRequest(c, requestID, "Meal plan ID is required")
	}

	err := h.mealPlanService.DeactivateMealPlan(ctx, mealPlanID)
	if err != nil {
		return h.handleMealPlanError(c, errHandler, requestID, err, c.Path(), "deactivate_meal_plan")
	}

	return response.Success(c, nil, "Meal plan deactivated successfully")
}

func (h *MealPlanHandler) BulkUpdateStatus(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 30*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.middleware.GetRequestID(c)

	var req struct {
		IDs      []string `json:"ids" validate:"required,min=1"`
		IsActive bool     `json:"is_active"`
	}

	if err := c.BodyParser(&req); err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "parse_request_body")
	}

	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(c, requestID, err, c.Path())
	}

	err := h.mealPlanService.BulkUpdateActiveStatus(ctx, req.IDs, req.IsActive)
	if err != nil {
		return h.handleMealPlanError(c, errHandler, requestID, err, c.Path(), "bulk_update_status")
	}

	action := "activated"
	if !req.IsActive {
		action = "deactivated"
	}

	return response.Success(c, nil, fmt.Sprintf("Meal plans %s successfully", action))
}

func (h *MealPlanHandler) GetMealPlanStats(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.middleware.GetRequestID(c)

	stats, err := h.mealPlanService.GetMealPlanStats(ctx)
	if err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "get_meal_plan_stats")
	}

	return response.Success(c, stats)
}

func (h *MealPlanHandler) handleMealPlanError(c *fiber.Ctx, errHandler *handlerutil.ErrorHandler, requestID string, err error, path, operation string) error {
	switch err {
	case meal_plans.ErrMealPlanNotFound:
		return errHandler.HandleNotFound(c, requestID, "Meal plan")
	case meal_plans.ErrMealPlanAlreadyExists, meal_plans.ErrMealPlanNameTaken:
		return response.Conflict(c, meal_plans.GetErrorMessage(err))
	case meal_plans.ErrMealPlanInactive, meal_plans.ErrInvalidPrice, meal_plans.ErrInvalidFeatures,
		meal_plans.ErrInvalidSortField, meal_plans.ErrInvalidImageURL, meal_plans.ErrInvalidPaginationParams:
		return errHandler.HandleBadRequest(c, requestID, meal_plans.GetErrorMessage(err))
	case meal_plans.ErrMealPlanHasSubscriptions:
		return response.UnprocessableEntity(c, meal_plans.GetErrorMessage(err))
	default:
		return errHandler.Handle(c, requestID, err, path, operation)
	}
}
