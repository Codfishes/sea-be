package handlerutil

import (
	"fmt"
	"sea-catering-backend/pkg/logger"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	"sea-catering-backend/pkg/response"
)

type ErrorHandler struct {
	logger *logger.Logger
}

func New(logger *logger.Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

func (e *ErrorHandler) Handle(c *fiber.Ctx, requestID string, err error, path, operation string) error {
	e.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"path":       path,
		"operation":  operation,
		"error":      err.Error(),
	}).Error("Handler error occurred")

	return response.InternalServerError(c, "An error occurred while processing your request")
}

func (e *ErrorHandler) HandleValidationError(c *fiber.Ctx, requestID string, err error, path string) error {
	var errors []response.ErrorDetail

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, validationErr := range validationErrors {
			errors = append(errors, response.ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: getValidationMessage(validationErr),
				Field:   validationErr.Field(),
				Value:   validationErr.Value(),
			})
		}
	} else {
		errors = append(errors, response.ErrorDetail{
			Code:    "VALIDATION_ERROR",
			Message: err.Error(),
		})
	}

	e.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"path":       path,
		"errors":     errors,
	}).Warn("Validation error occurred")

	return response.ValidationError(c, errors)
}

func (e *ErrorHandler) HandleSuccess(c *fiber.Ctx, statusCode int, data interface{}) error {
	return c.Status(statusCode).JSON(response.StandardResponse{
		Success: true,
		Data:    data,
	})
}

func (e *ErrorHandler) HandleUnauthorized(c *fiber.Ctx, requestID string, message string) error {
	e.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"message":    message,
	}).Warn("Unauthorized access attempt")

	return response.Unauthorized(c, message)
}

func (e *ErrorHandler) HandleNotFound(c *fiber.Ctx, requestID string, resource string) error {
	message := fmt.Sprintf("%s not found", resource)

	e.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"resource":   resource,
	}).Warn("Resource not found")

	return response.NotFound(c, message)
}

func (e *ErrorHandler) HandleBadRequest(c *fiber.Ctx, requestID string, message string) error {
	e.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"message":    message,
	}).Warn("Bad request")

	return response.BadRequest(c, message)
}

func (e *ErrorHandler) HandleForbidden(c *fiber.Ctx, requestID string, message string) error {
	e.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"message":    message,
	}).Warn("Forbidden access")

	return response.Forbidden(c, message)
}

func getValidationMessage(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, err.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, err.Param())
	case "phone_id":
		return fmt.Sprintf("%s must be a valid Indonesian phone number", field)
	case "strong_password":
		return fmt.Sprintf("%s must contain at least 8 characters with uppercase, lowercase, number, and special character", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
