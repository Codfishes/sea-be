package handler

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"sea-catering-backend/internal/api/testimonials"
	"sea-catering-backend/internal/api/testimonials/service"
	"sea-catering-backend/internal/middleware"
	"sea-catering-backend/pkg/context"
	"sea-catering-backend/pkg/handlerutil"
	"sea-catering-backend/pkg/logger"
)

type TestimonialHandler struct {
	testimonialService service.TestimonialService
	validator          *validator.Validate
	middleware         middleware.Interface
	logger             *logger.Logger
}

func NewTestimonialHandler(
	testimonialService service.TestimonialService,
	validator *validator.Validate,
	middleware middleware.Interface,
	logger *logger.Logger,
) *TestimonialHandler {
	return &TestimonialHandler{
		testimonialService: testimonialService,
		validator:          validator,
		middleware:         middleware,
		logger:             logger,
	}
}

func (h *TestimonialHandler) RegisterRoutes(router fiber.Router) {
	testimonialsGroup := router.Group("/testimonials")

	testimonialsGroup.Post("/", h.CreateTestimonial)
	testimonialsGroup.Get("/", h.GetApprovedTestimonials)

	admin := testimonialsGroup.Group("/admin", h.middleware.AdminMiddleware())
	admin.Get("/all", h.GetAllTestimonials)
	admin.Put("/:id/approve", h.ApproveTestimonial)
	admin.Put("/:id/reject", h.RejectTestimonial)
	admin.Delete("/:id", h.DeleteTestimonial)
}

func (h *TestimonialHandler) CreateTestimonial(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	var req testimonials.CreateTestimonialRequest
	if err := c.BodyParser(&req); err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "parse_request_body")
	}

	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(c, requestID, err, c.Path())
	}

	testimonial, err := h.testimonialService.CreateTestimonial(ctx, req)
	if err != nil {
		return h.handleTestimonialError(c, errHandler, requestID, err, c.Path(), "create_testimonial")
	}

	return errHandler.HandleSuccess(c, fiber.StatusCreated, fiber.Map{
		"message":     "Testimonial submitted successfully and is pending approval",
		"testimonial": testimonial,
	})
}

func (h *TestimonialHandler) GetApprovedTestimonials(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	testimonialsList, err := h.testimonialService.GetApprovedTestimonials(ctx)
	if err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "get_approved_testimonials")
	}

	return errHandler.HandleSuccess(c, fiber.StatusOK, testimonialsList)
}

func (h *TestimonialHandler) GetAllTestimonials(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	testimonialsList, err := h.testimonialService.GetAllTestimonials(ctx)
	if err != nil {
		return errHandler.Handle(c, requestID, err, c.Path(), "get_all_testimonials")
	}

	return errHandler.HandleSuccess(c, fiber.StatusOK, testimonialsList)
}

func (h *TestimonialHandler) ApproveTestimonial(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	testimonialID := c.Params("id")
	if testimonialID == "" {
		return errHandler.HandleBadRequest(c, requestID, "Testimonial ID is required")
	}

	err := h.testimonialService.ApproveTestimonial(ctx, testimonialID)
	if err != nil {
		return h.handleTestimonialError(c, errHandler, requestID, err, c.Path(), "approve_testimonial")
	}

	return errHandler.HandleSuccess(c, fiber.StatusOK, fiber.Map{
		"message": "Testimonial approved successfully",
	})
}

func (h *TestimonialHandler) RejectTestimonial(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	testimonialID := c.Params("id")
	if testimonialID == "" {
		return errHandler.HandleBadRequest(c, requestID, "Testimonial ID is required")
	}

	err := h.testimonialService.RejectTestimonial(ctx, testimonialID)
	if err != nil {
		return h.handleTestimonialError(c, errHandler, requestID, err, c.Path(), "reject_testimonial")
	}

	return errHandler.HandleSuccess(c, fiber.StatusOK, fiber.Map{
		"message": "Testimonial rejected successfully",
	})
}

func (h *TestimonialHandler) DeleteTestimonial(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.FromFiberContext(c), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.logger)
	requestID := h.getRequestID(c)

	testimonialID := c.Params("id")
	if testimonialID == "" {
		return errHandler.HandleBadRequest(c, requestID, "Testimonial ID is required")
	}

	err := h.testimonialService.DeleteTestimonial(ctx, testimonialID)
	if err != nil {
		return h.handleTestimonialError(c, errHandler, requestID, err, c.Path(), "delete_testimonial")
	}

	return errHandler.HandleSuccess(c, fiber.StatusOK, fiber.Map{
		"message": "Testimonial deleted successfully",
	})
}

func (h *TestimonialHandler) getRequestID(c *fiber.Ctx) string {
	if requestID := c.Locals("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return c.Get("X-Request-ID", "unknown")
}

func (h *TestimonialHandler) handleTestimonialError(c *fiber.Ctx, errHandler *handlerutil.ErrorHandler, requestID string, err error, path, operation string) error {
	switch err {
	case testimonials.ErrTestimonialNotFound:
		return errHandler.HandleNotFound(c, requestID, "Testimonial")
	case testimonials.ErrInvalidRating:
		return errHandler.HandleBadRequest(c, requestID, "Rating must be between 1 and 5")
	case testimonials.ErrUnauthorizedAccess:
		return errHandler.HandleForbidden(c, requestID, "Unauthorized access")
	default:
		return errHandler.Handle(c, requestID, err, path, operation)
	}
}
