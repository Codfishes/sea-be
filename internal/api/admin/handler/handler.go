package handler

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"sea-catering-backend/internal/api/admin"
	"sea-catering-backend/internal/api/admin/service"
	"sea-catering-backend/internal/middleware"
	"sea-catering-backend/pkg/context"
	"sea-catering-backend/pkg/handlerutil"
	"sea-catering-backend/pkg/logger"
)

type AdminHandler struct {
	adminService service.AdminService
	validator    *validator.Validate
	middleware   middleware.Interface
	logger       *logger.Logger
}

func NewAdminHandler(
	adminService service.AdminService,
	validator *validator.Validate,
	middleware middleware.Interface,
	logger *logger.Logger,
) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
		validator:    validator,
		middleware:   middleware,
		logger:       logger,
	}
}

func (h *AdminHandler) RegisterRoutes(router fiber.Router) {
	adminGroup := router.Group("/admin")

	adminGroup.Post("/login", h.AdminLogin)

	protected := adminGroup.Use(h.middleware.AdminMiddleware())
	protected.Get("/dashboard", h.GetDashboardStats)
	protected.Post("/dashboard/filter", h.GetDashboardStatsWithFilter)

	protected.Put("/testimonials/:id/approve", h.ApproveTestimonial)
	protected.Put("/testimonials/:id/reject", h.RejectTestimonial)

	protected.Get("/users", h.GetAllUsers)
	protected.Get("/users/:id", h.GetUserByID)
	protected.Put("/users/:id/status", h.UpdateUserStatus)
	protected.Delete("/users/:id", h.DeleteUser)

	subscriptionGroup := router.Group("/subscriptions/admin")
	subscriptionProtected := subscriptionGroup.Use(h.middleware.AdminMiddleware())

	subscriptionProtected.Get("/search", h.SearchSubscriptions)
	subscriptionProtected.Put("/:id/force-cancel", h.ForceCancelSubscription)
}

func (h *AdminHandler) AdminLogin(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	var req admin.AdminLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "parse_request_body")
	}

	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(c, requestID, err, c.Path())
	}

	response, err := h.adminService.AdminLogin(ctx, req)
	if err != nil {
		return h.handleAdminError(c, errHandler, requestID, err, c.Path(), "admin_login")
	}

	return errHandler.HandleSuccess(c, fiber.StatusOK, response)
}

func (h *AdminHandler) GetDashboardStats(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	stats, err := h.adminService.GetDashboardStats(ctx)
	if err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "get_dashboard_stats")
	}

	return errHandler.HandleSuccess(c, fiber.StatusOK, stats)
}

func (h *AdminHandler) GetDashboardStatsWithFilter(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	var filter admin.DateRangeFilter
	if err := c.BodyParser(&filter); err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "parse_request_body")
	}

	if err := h.validator.Struct(filter); err != nil {
		return errHandler.HandleValidationError(c, requestID, err, c.Path())
	}

	stats, err := h.adminService.GetDashboardStatsWithFilter(ctx, filter.StartDate, filter.EndDate)
	if err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "get_dashboard_stats_with_filter")
	}

	return errHandler.HandleSuccess(c, fiber.StatusOK, stats)
}

func (h *AdminHandler) ApproveTestimonial(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	testimonialID := c.Params("id")
	if testimonialID == "" {
		return errHandler.HandleBadRequest(c, requestID, "Testimonial ID is required")
	}

	err := h.adminService.ApproveTestimonial(ctx, testimonialID)
	if err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "approve_testimonial")
	}

	return errHandler.HandleSuccess(c, fiber.StatusOK, fiber.Map{
		"message": "Testimonial approved successfully",
	})
}

func (h *AdminHandler) RejectTestimonial(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	testimonialID := c.Params("id")
	if testimonialID == "" {
		return errHandler.HandleBadRequest(c, requestID, "Testimonial ID is required")
	}

	err := h.adminService.RejectTestimonial(ctx, testimonialID)
	if err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "reject_testimonial")
	}

	return errHandler.HandleSuccess(c, fiber.StatusOK, fiber.Map{
		"message": "Testimonial rejected successfully",
	})
}

func (h *AdminHandler) getRequestID(c *fiber.Ctx) string {
	if requestID := c.Locals("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return c.Get("X-Request-ID", "unknown")
}

func (h *AdminHandler) handleAdminError(c *fiber.Ctx, errHandler *handlerutil.ErrorHandler, requestID string, err error, path, operation string) error {
	switch err {
	case admin.ErrAdminNotFound:
		return errHandler.HandleNotFound(c, requestID, "Admin")
	case admin.ErrInvalidCredentials:
		return errHandler.HandleUnauthorized(c, requestID, "Invalid credentials")
	case admin.ErrAdminAlreadyExists:
		return errHandler.HandleBadRequest(c, requestID, "Admin already exists")
	case admin.ErrUnauthorizedAccess:
		return errHandler.HandleForbidden(c, requestID, "Unauthorized access")
	case admin.ErrInvalidRole:
		return errHandler.HandleBadRequest(c, requestID, "Invalid admin role")
	case admin.ErrInsufficientRights:
		return errHandler.HandleForbidden(c, requestID, "Insufficient admin rights")
	default:
		return errHandler.Handle(c, requestID, err, path, operation)
	}
}

func (h *AdminHandler) GetAllUsers(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	var req admin.UserListRequest
	if err := c.QueryParser(&req); err != nil {
		return errHandler.HandleBadRequest(c, requestID, "Invalid query parameters")
	}

	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(c, requestID, err, c.Path())
	}

	result, err := h.adminService.GetAllUsers(ctx, req)
	if err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "get_all_users")
	}

	return errHandler.HandleSuccess(c, fiber.StatusOK, result)
}

func (h *AdminHandler) GetUserByID(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	userID := c.Params("id")
	if userID == "" {
		return errHandler.HandleBadRequest(c, requestID, "User ID is required")
	}

	user, err := h.adminService.GetUserByID(ctx, userID)
	if err != nil {
		if err.Error() == "user not found" {
			return errHandler.HandleNotFound(c, requestID, "User")
		}
		return errHandler.Handle(c, requestID, err, c.Path(), "get_user_by_id")
	}

	return errHandler.HandleSuccess(c, fiber.StatusOK, user)
}

func (h *AdminHandler) UpdateUserStatus(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	userID := c.Params("id")
	if userID == "" {
		return errHandler.HandleBadRequest(c, requestID, "User ID is required")
	}

	var req admin.UpdateUserStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "parse_request_body")
	}

	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(c, requestID, err, c.Path())
	}

	err := h.adminService.UpdateUserStatus(ctx, userID, req)
	if err != nil {
		if err.Error() == "user not found" {
			return errHandler.HandleNotFound(c, requestID, "User")
		}
		return errHandler.Handle(c, requestID, err, c.Path(), "update_user_status")
	}

	action := "activated"
	if !req.IsActive {
		action = "deactivated"
	}

	return errHandler.HandleSuccess(c, fiber.StatusOK, fiber.Map{
		"message": "User " + action + " successfully",
	})
}

func (h *AdminHandler) DeleteUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	userID := c.Params("id")
	if userID == "" {
		return errHandler.HandleBadRequest(c, requestID, "User ID is required")
	}

	err := h.adminService.DeleteUser(ctx, userID)
	if err != nil {
		if err.Error() == "user not found" {
			return errHandler.HandleNotFound(c, requestID, "User")
		}
		if err.Error() == "cannot delete user with active subscriptions" ||
			(len(err.Error()) > 30 && err.Error()[:30] == "cannot delete user with") {
			return errHandler.HandleBadRequest(c, requestID, err.Error())
		}
		return errHandler.Handle(c, requestID, err, c.Path(), "delete_user")
	}

	return errHandler.HandleSuccess(c, fiber.StatusOK, fiber.Map{
		"message": "User deleted successfully",
	})
}

func (h *AdminHandler) SearchSubscriptions(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 15*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	var req admin.SubscriptionSearchRequest
	if err := c.QueryParser(&req); err != nil {
		return errHandler.HandleBadRequest(c, requestID, "Invalid query parameters")
	}

	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(c, requestID, err, c.Path())
	}

	result, err := h.adminService.SearchSubscriptions(ctx, req)
	if err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "search_subscriptions")
	}

	return errHandler.HandleSuccess(c, fiber.StatusOK, result)
}

func (h *AdminHandler) ForceCancelSubscription(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 30*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	subscriptionID := c.Params("id")
	if subscriptionID == "" {
		return errHandler.HandleBadRequest(c, requestID, "Subscription ID is required")
	}

	var req admin.ForceCancelSubscriptionRequest
	if err := c.BodyParser(&req); err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "parse_request_body")
	}

	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(c, requestID, err, c.Path())
	}

	result, err := h.adminService.ForceCancelSubscription(ctx, subscriptionID, req)
	if err != nil {
		if err.Error() == "subscription not found" {
			return errHandler.HandleNotFound(c, requestID, "Subscription")
		}
		if err.Error() == "subscription is already cancelled" {
			return errHandler.HandleBadRequest(c, requestID, err.Error())
		}
		return errHandler.Handle(c, requestID, err, c.Path(), "force_cancel_subscription")
	}

	return errHandler.HandleSuccess(c, fiber.StatusOK, result)
}
