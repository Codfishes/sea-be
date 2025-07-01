package response

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type StandardResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     interface{} `json:"error,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

type Meta struct {
	Page       int  `json:"page,omitempty"`
	Limit      int  `json:"limit,omitempty"`
	Total      int  `json:"total,omitempty"`
	TotalPages int  `json:"total_pages,omitempty"`
	HasNext    bool `json:"has_next,omitempty"`
	HasPrev    bool `json:"has_prev,omitempty"`
}

type ErrorDetail struct {
	Code    string      `json:"code,omitempty"`
	Message string      `json:"message"`
	Field   string      `json:"field,omitempty"`
	Value   interface{} `json:"value,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

type ValidationErrors struct {
	Message string        `json:"message"`
	Errors  []ErrorDetail `json:"errors"`
}

func Success(c *fiber.Ctx, data interface{}, message ...string) error {
	response := StandardResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	}

	if len(message) > 0 {
		response.Message = message[0]
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func Created(c *fiber.Ctx, data interface{}, message ...string) error {
	response := StandardResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	}

	if len(message) > 0 {
		response.Message = message[0]
	} else {
		response.Message = "Resource created successfully"
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

func Updated(c *fiber.Ctx, data interface{}, message ...string) error {
	response := StandardResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	}

	if len(message) > 0 {
		response.Message = message[0]
	} else {
		response.Message = "Resource updated successfully"
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func Deleted(c *fiber.Ctx, message ...string) error {
	response := StandardResponse{
		Success:   true,
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	}

	if len(message) > 0 {
		response.Message = message[0]
	} else {
		response.Message = "Resource deleted successfully"
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func Paginated(c *fiber.Ctx, data interface{}, meta *Meta, message ...string) error {
	response := StandardResponse{
		Success:   true,
		Data:      data,
		Meta:      meta,
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	}

	if len(message) > 0 {
		response.Message = message[0]
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func Error(c *fiber.Ctx, statusCode int, message string, details ...interface{}) error {
	response := StandardResponse{
		Success:   false,
		Message:   message,
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	}

	if len(details) > 0 {
		response.Error = details[0]
	}

	return c.Status(statusCode).JSON(response)
}

func BadRequest(c *fiber.Ctx, message string, details ...interface{}) error {
	return Error(c, fiber.StatusBadRequest, message, details...)
}

func Unauthorized(c *fiber.Ctx, message ...string) error {
	msg := "Unauthorized"
	if len(message) > 0 {
		msg = message[0]
	}
	return Error(c, fiber.StatusUnauthorized, msg)
}

func Forbidden(c *fiber.Ctx, message ...string) error {
	msg := "Forbidden"
	if len(message) > 0 {
		msg = message[0]
	}
	return Error(c, fiber.StatusForbidden, msg)
}

func NotFound(c *fiber.Ctx, message ...string) error {
	msg := "Resource not found"
	if len(message) > 0 {
		msg = message[0]
	}
	return Error(c, fiber.StatusNotFound, msg)
}

func Conflict(c *fiber.Ctx, message string, details ...interface{}) error {
	return Error(c, fiber.StatusConflict, message, details...)
}

func UnprocessableEntity(c *fiber.Ctx, message string, details ...interface{}) error {
	return Error(c, fiber.StatusUnprocessableEntity, message, details...)
}

func TooManyRequests(c *fiber.Ctx, message ...string) error {
	msg := "Too many requests"
	if len(message) > 0 {
		msg = message[0]
	}
	return Error(c, fiber.StatusTooManyRequests, msg)
}

func InternalServerError(c *fiber.Ctx, message ...string) error {
	msg := "Internal server error"
	if len(message) > 0 {
		msg = message[0]
	}
	return Error(c, fiber.StatusInternalServerError, msg)
}

func ValidationError(c *fiber.Ctx, errors []ErrorDetail) error {
	validationErrors := ValidationErrors{
		Message: "Validation failed",
		Errors:  errors,
	}

	return Error(c, fiber.StatusBadRequest, "Validation failed", validationErrors)
}

func Custom(c *fiber.Ctx, statusCode int, success bool, message string, data interface{}, meta *Meta) error {
	response := StandardResponse{
		Success:   success,
		Message:   message,
		Data:      data,
		Meta:      meta,
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	}

	return c.Status(statusCode).JSON(response)
}

func NewMeta(page, limit, total int) *Meta {
	totalPages := 0
	if limit > 0 {
		totalPages = (total + limit - 1) / limit
	}

	return &Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

func NewErrorDetail(code, message, field string, value interface{}) ErrorDetail {
	return ErrorDetail{
		Code:    code,
		Message: message,
		Field:   field,
		Value:   value,
	}
}

func getRequestID(c *fiber.Ctx) string {
	if requestID := c.Locals("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			return id
		}
	}

	if requestID := c.Get("X-Request-ID"); requestID != "" {
		return requestID
	}

	return ""
}

const (
	StatusOK                  = fiber.StatusOK
	StatusCreated             = fiber.StatusCreated
	StatusNoContent           = fiber.StatusNoContent
	StatusBadRequest          = fiber.StatusBadRequest
	StatusUnauthorized        = fiber.StatusUnauthorized
	StatusForbidden           = fiber.StatusForbidden
	StatusNotFound            = fiber.StatusNotFound
	StatusConflict            = fiber.StatusConflict
	StatusUnprocessableEntity = fiber.StatusUnprocessableEntity
	StatusTooManyRequests     = fiber.StatusTooManyRequests
	StatusInternalServerError = fiber.StatusInternalServerError
)

const (
	ErrorCodeValidation         = "VALIDATION_ERROR"
	ErrorCodeNotFound           = "NOT_FOUND"
	ErrorCodeUnauthorized       = "UNAUTHORIZED"
	ErrorCodeForbidden          = "FORBIDDEN"
	ErrorCodeConflict           = "CONFLICT"
	ErrorCodeInternalError      = "INTERNAL_ERROR"
	ErrorCodeInvalidInput       = "INVALID_INPUT"
	ErrorCodeDuplicateEntry     = "DUPLICATE_ENTRY"
	ErrorCodeRateLimit          = "RATE_LIMIT_EXCEEDED"
	ErrorCodeServiceDown        = "SERVICE_UNAVAILABLE"
	ErrorCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrorCodeExpiredToken       = "EXPIRED_TOKEN"
	ErrorCodeInvalidToken       = "INVALID_TOKEN"
	ErrorCodePermissionDenied   = "PERMISSION_DENIED"
	ErrorCodeResourceExhausted  = "RESOURCE_EXHAUSTED"
)

const (
	ErrorCodeInvalidMealPlan          = "INVALID_MEAL_PLAN"
	ErrorCodeInvalidMealType          = "INVALID_MEAL_TYPE"
	ErrorCodeInvalidDeliveryDay       = "INVALID_DELIVERY_DAY"
	ErrorCodeSubscriptionExists       = "SUBSCRIPTION_EXISTS"
	ErrorCodeSubscriptionNotFound     = "SUBSCRIPTION_NOT_FOUND"
	ErrorCodePaymentFailed            = "PAYMENT_FAILED"
	ErrorCodeInvalidPaymentMethod     = "INVALID_PAYMENT_METHOD"
	ErrorCodeDeliveryAreaNotSupported = "DELIVERY_AREA_NOT_SUPPORTED"
	ErrorCodeOrderNotFound            = "ORDER_NOT_FOUND"
	ErrorCodeInvalidOrderStatus       = "INVALID_ORDER_STATUS"
	ErrorCodeTestimonialExists        = "TESTIMONIAL_EXISTS"
	ErrorCodeInvalidRating            = "INVALID_RATING"
)

func SubscriptionCreated(c *fiber.Ctx, subscription interface{}) error {
	return Created(c, subscription, "Subscription created successfully")
}

func SubscriptionUpdated(c *fiber.Ctx, subscription interface{}) error {
	return Updated(c, subscription, "Subscription updated successfully")
}

func SubscriptionCancelled(c *fiber.Ctx) error {
	return Success(c, nil, "Subscription cancelled successfully")
}

func OrderCreated(c *fiber.Ctx, order interface{}) error {
	return Created(c, order, "Order created successfully")
}

func PaymentProcessed(c *fiber.Ctx, payment interface{}) error {
	return Success(c, payment, "Payment processed successfully")
}

func TestimonialCreated(c *fiber.Ctx, testimonial interface{}) error {
	return Created(c, testimonial, "Testimonial submitted successfully")
}

func InvalidMealPlan(c *fiber.Ctx, details ...interface{}) error {
	return BadRequest(c, "Invalid meal plan selected", details...)
}

func InvalidMealType(c *fiber.Ctx, details ...interface{}) error {
	return BadRequest(c, "Invalid meal type selected", details...)
}

func InvalidDeliveryDay(c *fiber.Ctx, details ...interface{}) error {
	return BadRequest(c, "Invalid delivery day selected", details...)
}

func SubscriptionNotFound(c *fiber.Ctx) error {
	return NotFound(c, "Subscription not found")
}

func PaymentFailed(c *fiber.Ctx, details ...interface{}) error {
	return BadRequest(c, "Payment processing failed", details...)
}

func DeliveryAreaNotSupported(c *fiber.Ctx) error {
	return BadRequest(c, "Delivery area not supported")
}

func AdminOnly(c *fiber.Ctx) error {
	return Forbidden(c, "Admin access required")
}

func SubscriptionRequired(c *fiber.Ctx) error {
	return Forbidden(c, "Active subscription required")
}
