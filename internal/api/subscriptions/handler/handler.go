package handler

import (
	contexts "context"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"sea-catering-backend/internal/api/subscriptions"
	"sea-catering-backend/internal/api/subscriptions/service"
	"sea-catering-backend/internal/middleware"
	"sea-catering-backend/pkg/context"
	"sea-catering-backend/pkg/handlerutil"
	"sea-catering-backend/pkg/jwt"
	"sea-catering-backend/pkg/logger"
	"strconv"
	"time"
)

type SubscriptionHandler struct {
	subscriptionService service.SubscriptionService
	validator           *validator.Validate
	middleware          middleware.Interface
	logger              *logger.Logger
}

func NewSubscriptionHandler(
	subscriptionService service.SubscriptionService,
	validator *validator.Validate,
	middleware middleware.Interface,
	logger *logger.Logger,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
		validator:           validator,
		middleware:          middleware,
		logger:              logger,
	}
}

func (h *SubscriptionHandler) RegisterRoutes(router fiber.Router) {
	subs := router.Group("/subscriptions")

	protected := subs.Use(h.middleware.AuthMiddleware())
	protected.Post("/", h.CreateSubscription)
	protected.Get("/my", h.GetMySubscriptions)
	protected.Get("/:id", h.GetSubscription)
	protected.Put("/:id", h.UpdateSubscription)
	protected.Put("/:id/pause", h.PauseSubscription)
	protected.Put("/:id/resume", h.ResumeSubscription)
	protected.Put("/:id/reactivate", h.ReactivateSubscription)
	protected.Delete("/:id", h.CancelSubscription)

	admin := subs.Group("/admin", h.middleware.AdminMiddleware())
	admin.Get("/stats", h.GetSubscriptionStats)
	admin.Get("/all", h.GetAllSubscriptions)
	admin.Post("/process-expired", h.ProcessExpiredPauses)
}

func (h *SubscriptionHandler) CreateSubscription(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	userID, err := jwt.GetUserID(c)
	if err != nil {
		return errHandler.HandleUnauthorized(c, requestID, "Authentication required")
	}

	var req subscriptions.CreateSubscriptionRequest
	if err := c.BodyParser(&req); err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "parse_request_body")
	}

	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(c, requestID, err, c.Path())
	}

	ctx = contexts.WithValue(ctx, "user_id", userID)

	subscription, err := h.subscriptionService.CreateSubscription(ctx, req)
	if err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "create_subscription")
	}

	return errHandler.HandleSuccess(c, fiber.StatusCreated, subscription)
}

func (h *SubscriptionHandler) GetMySubscriptions(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	userID, err := jwt.GetUserID(c)
	if err != nil {
		h.logger.Error("Failed to get user ID from JWT", logger.Fields{
			"error":      err.Error(),
			"request_id": requestID,
		})
		return errHandler.HandleUnauthorized(c, requestID, "Authentication required")
	}

	h.logger.Info("Getting user subscriptions", logger.Fields{
		"user_id":    userID,
		"request_id": requestID,
	})

	subscriptions, err := h.subscriptionService.GetUserSubscriptions(ctx, userID)
	if err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "get_user_subscriptions")
	}

	return errHandler.HandleSuccess(c, fiber.StatusOK, subscriptions)
}

func (h *SubscriptionHandler) GetSubscription(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	subscriptionID := c.Params("id")
	userID, err := jwt.GetUserID(c)
	if err != nil {
		return errHandler.HandleUnauthorized(c, requestID, "Unauthorized")
	}

	subscription, err := h.subscriptionService.GetSubscriptionByID(ctx, subscriptionID)
	if err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "get_subscription")
	}

	if subscription.UserID != userID {
		return errHandler.HandleForbidden(c, requestID, "Access denied")
	}

	return errHandler.HandleSuccess(c, fiber.StatusOK, subscription)
}

func (h *SubscriptionHandler) UpdateSubscription(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 30*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	subscriptionID := c.Params("id")
	userID, err := jwt.GetUserID(c)
	if err != nil {
		return errHandler.HandleUnauthorized(c, requestID, "Unauthorized")
	}

	var req subscriptions.CreateSubscriptionRequest
	if err := c.BodyParser(&req); err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "parse_request_body")
	}

	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(c, requestID, err, c.Path())
	}

	subscription, err := h.subscriptionService.UpdateSubscription(ctx, subscriptionID, userID, req)
	if err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "update_subscription")
	}

	return errHandler.HandleSuccess(c, fiber.StatusOK, subscription)
}

func (h *SubscriptionHandler) PauseSubscription(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	subscriptionID := c.Params("id")
	userID, err := jwt.GetUserID(c)
	if err != nil {
		return errHandler.HandleUnauthorized(c, requestID, "Unauthorized")
	}

	var req subscriptions.PauseSubscriptionRequest
	if err := c.BodyParser(&req); err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "parse_request_body")
	}

	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(c, requestID, err, c.Path())
	}

	err = h.subscriptionService.PauseSubscription(ctx, subscriptionID, userID, req.StartDate, req.EndDate)
	if err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "pause_subscription")
	}

	return errHandler.HandleSuccess(c, fiber.StatusOK, fiber.Map{
		"message": "Subscription paused successfully",
	})
}

func (h *SubscriptionHandler) ResumeSubscription(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	subscriptionID := c.Params("id")
	userID, err := jwt.GetUserID(c)
	if err != nil {
		return errHandler.HandleUnauthorized(c, requestID, "Unauthorized")
	}

	err = h.subscriptionService.ResumeSubscription(ctx, subscriptionID, userID)
	if err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "resume_subscription")
	}

	return errHandler.HandleSuccess(c, fiber.StatusOK, fiber.Map{
		"message": "Subscription resumed successfully",
	})
}

func (h *SubscriptionHandler) ReactivateSubscription(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 30*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	subscriptionID := c.Params("id")
	if subscriptionID == "" {
		return errHandler.HandleBadRequest(c, requestID, "Subscription ID is required")
	}

	userID, err := jwt.GetUserID(c)
	if err != nil {
		return errHandler.HandleUnauthorized(c, requestID, "Authentication required")
	}

	h.logger.Info("Processing subscription reactivation request", logger.Fields{
		"subscription_id": subscriptionID,
		"user_id":         userID,
		"request_id":      requestID,
		"ip":              c.IP(),
		"user_agent":      c.Get("User-Agent"),
	})

	_, err = h.subscriptionService.ReactivateSubscription(ctx, subscriptionID, userID)
	if err != nil {
		h.logger.Error("Subscription reactivation failed", logger.Fields{
			"error":           err.Error(),
			"subscription_id": subscriptionID,
			"user_id":         userID,
			"request_id":      requestID,
		})

		switch err.Error() {
		case "failed to get subscription: subscription not found":
			return errHandler.HandleNotFound(c, requestID, "Subscription")
		case "only cancelled subscriptions can be reactivated, current status: active":
			return errHandler.HandleBadRequest(c, requestID, "Subscription is already active")
		case "only cancelled subscriptions can be reactivated, current status: paused":
			return errHandler.HandleBadRequest(c, requestID, "Cannot reactivate paused subscription, please resume instead")
		default:
			if err == subscriptions.ErrUnauthorizedAccess {
				return errHandler.HandleForbidden(c, requestID, "Access denied")
			}

			return errHandler.Handle(c, requestID, err, c.Path(), "reactivate_subscription")
		}
	}

	h.logger.Info("Subscription reactivated successfully", logger.Fields{
		"subscription_id": subscriptionID,
		"user_id":         userID,
		"request_id":      requestID,
	})

	return errHandler.HandleSuccess(c, fiber.StatusOK, fiber.Map{
		"success": true,
		"message": "Subscription reactivated successfully. Please complete the payment to activate your subscription.",
		"next_steps": []string{
			"Complete the payment using the provided payment URL",
			"Your subscription will be activated upon successful payment",
			"You will receive confirmation via email once activated",
		},
	})
}

func (h *SubscriptionHandler) CancelSubscription(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	subscriptionID := c.Params("id")
	userID, err := jwt.GetUserID(c)
	if err != nil {
		return errHandler.HandleUnauthorized(c, requestID, "Unauthorized")
	}

	err = h.subscriptionService.CancelSubscription(ctx, subscriptionID, userID)
	if err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "cancel_subscription")
	}

	return errHandler.HandleSuccess(c, fiber.StatusOK, fiber.Map{
		"message": "Subscription cancelled successfully",
	})
}

func (h *SubscriptionHandler) GetSubscriptionStats(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return errHandler.HandleValidationError(c, requestID, err, c.Path())
		}
	} else {

		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return errHandler.HandleValidationError(c, requestID, err, c.Path())
		}
	} else {

		startOfMonth := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, startDate.Location())
		endDate = startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)
	}

	h.logger.Info("Getting subscription stats", logger.Fields{
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
		"request_id": requestID,
	})

	stats, err := h.subscriptionService.GetSubscriptionStats(ctx, startDate, endDate)
	if err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "get_subscription_stats")
	}

	response := fiber.Map{
		"period": fiber.Map{
			"start_date": startDate.Format("2006-01-02"),
			"end_date":   endDate.Format("2006-01-02"),
		},
		"subscriptions": fiber.Map{
			"total":               stats.TotalSubscriptions,
			"active":              stats.ActiveSubscriptions,
			"paused":              stats.PausedSubscriptions,
			"cancelled":           stats.CancelledSubscriptions,
			"new_subscriptions":   stats.NewSubscriptions,
			"reactivations":       stats.Reactivations,
			"subscription_growth": calculateSubscriptionGrowth(stats),
		},
		"revenue": fiber.Map{
			"monthly_revenue": stats.MonthlyRevenue,
		},
		"breakdown": fiber.Map{
			"subscriptions_by_plan": stats.SubscriptionsByPlan,
		},
		"metadata": fiber.Map{
			"generated_at": time.Now().Format(time.RFC3339),
			"period_days":  int(endDate.Sub(startDate).Hours() / 24),
		},
	}

	h.logger.Info("Subscription stats retrieved successfully", logger.Fields{
		"total_subscriptions":  stats.TotalSubscriptions,
		"active_subscriptions": stats.ActiveSubscriptions,
		"new_subscriptions":    stats.NewSubscriptions,
		"reactivations":        stats.Reactivations,
		"request_id":           requestID,
	})

	return errHandler.HandleSuccess(c, fiber.StatusOK, response)
}

func (h *SubscriptionHandler) GetAllSubscriptions(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	subscriptions, err := h.subscriptionService.GetUserSubscriptions(ctx, "")
	if err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "get_all_subscriptions")
	}

	return errHandler.HandleSuccess(c, fiber.StatusOK, subscriptions)
}

func (h *SubscriptionHandler) ProcessExpiredPauses(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 30*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	err := h.subscriptionService.ProcessExpiredPauses(ctx)
	if err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "process_expired_pauses")
	}

	return errHandler.HandleSuccess(c, fiber.StatusOK, fiber.Map{
		"message": "Expired pauses processed successfully",
	})
}

func (h *SubscriptionHandler) getRequestID(c *fiber.Ctx) string {
	if requestID := c.Locals("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return c.Get("X-Request-ID", "unknown")
}

func calculateSubscriptionGrowth(stats *subscriptions.SubscriptionStatsResponse) float64 {
	if stats.TotalSubscriptions == 0 {
		return 0.0
	}

	growth := float64(stats.NewSubscriptions+stats.Reactivations) / float64(stats.TotalSubscriptions) * 100

	return float64(int(growth*100)) / 100
}
