package handler

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"sea-catering-backend/internal/api/auth"
	"sea-catering-backend/internal/api/auth/service"
	"sea-catering-backend/pkg/jwt"
	"sea-catering-backend/pkg/logger"
	"sea-catering-backend/pkg/response"
)

type AuthHandler struct {
	authService service.AuthService
	validator   *validator.Validate
	logger      *logger.Logger
}

func NewAuthHandler(
	authService service.AuthService,
	validator *validator.Validate,
	logger *logger.Logger,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator,
		logger:      logger,
	}
}

func (h *AuthHandler) RegisterRoutes(router fiber.Router) {
	authGroup := router.Group("/auth")

	authGroup.Post("/register", h.Register)
	authGroup.Post("/login", h.Login)
	authGroup.Post("/forgot-password", h.ForgotPassword)
	authGroup.Post("/reset-password", h.ResetPassword)
	authGroup.Post("/send-otp", h.SendOTP)
	authGroup.Post("/verify-otp", h.VerifyOTP)

	userGroup := router.Group("/user")
	userGroup.Get("/profile", h.GetProfile)
	userGroup.Put("/profile", h.UpdateProfile)
	userGroup.Post("/change-password", h.ChangePassword)
	userGroup.Post("/profile/image", h.UploadProfileImage)
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {

	h.logger.Info("=== REGISTER REQUEST DEBUG ===", logger.Fields{
		"method":       c.Method(),
		"path":         c.Path(),
		"content_type": c.Get("Content-Type"),
		"body_length":  len(c.Body()),
		"headers":      c.GetReqHeaders(),
	})

	rawBody := c.Body()
	h.logger.Info("Raw request body", logger.Fields{
		"body":        string(rawBody),
		"body_length": len(rawBody),
	})

	if len(rawBody) == 0 {
		h.logger.Error("Request body is empty")
		return response.BadRequest(c, "Request body is required")
	}

	var req auth.RegisterRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		h.logger.Error("Manual JSON unmarshal failed", logger.Fields{
			"error": err.Error(),
			"body":  string(rawBody),
		})
		return response.BadRequest(c, "Invalid JSON format")
	}

	h.logger.Info("Manual parsing successful", logger.Fields{
		"name":  req.Name,
		"phone": req.Phone,
		"email": req.Email,
	})

	var req2 auth.RegisterRequest
	if err := c.BodyParser(&req2); err != nil {
		h.logger.Error("Fiber BodyParser failed", logger.Fields{
			"error": err.Error(),
		})
		return response.BadRequest(c, "Failed to parse request body")
	}

	h.logger.Info("Fiber BodyParser result", logger.Fields{
		"name":  req2.Name,
		"phone": req2.Phone,
		"email": req2.Email,
	})

	finalReq := req
	if req2.Name != "" {
		finalReq = req2
	}

	fmt.Println("Final Register request body:", finalReq)

	if err := h.validator.Struct(finalReq); err != nil {
		h.logger.Error("Validation failed", logger.Fields{
			"error": err.Error(),
		})
		return h.handleValidationError(c, err)
	}

	result, err := h.authService.Register(c.Context(), finalReq)
	if err != nil {
		return h.handleError(c, err)
	}

	return response.Created(c, result, "User registered successfully")
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req auth.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := h.validator.Struct(req); err != nil {
		return h.handleValidationError(c, err)
	}

	result, err := h.authService.Login(c.Context(), req)
	if err != nil {
		return h.handleError(c, err)
	}

	return response.Success(c, result, "Login successful")
}

func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		return response.Unauthorized(c)
	}

	result, err := h.authService.GetProfile(c.Context(), userID)
	if err != nil {
		return h.handleError(c, err)
	}

	return response.Success(c, result)
}

func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		return response.Unauthorized(c)
	}

	var req auth.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := h.validator.Struct(req); err != nil {
		return h.handleValidationError(c, err)
	}

	result, err := h.authService.UpdateProfile(c.Context(), userID, req)
	if err != nil {
		return h.handleError(c, err)
	}

	return response.Updated(c, result, "Profile updated successfully")
}

func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		return response.Unauthorized(c)
	}

	var req auth.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := h.validator.Struct(req); err != nil {
		return h.handleValidationError(c, err)
	}

	err = h.authService.ChangePassword(c.Context(), userID, req)
	if err != nil {
		return h.handleError(c, err)
	}

	return response.Success(c, nil, "Password changed successfully")
}

func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req auth.ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := h.validator.Struct(req); err != nil {
		return h.handleValidationError(c, err)
	}

	err := h.authService.ForgotPassword(c.Context(), req)
	if err != nil {
		return h.handleError(c, err)
	}

	return response.Success(c, nil, "OTP sent successfully")
}

func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req auth.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := h.validator.Struct(req); err != nil {
		return h.handleValidationError(c, err)
	}

	err := h.authService.ResetPassword(c.Context(), req)
	if err != nil {
		return h.handleError(c, err)
	}

	return response.Success(c, nil, "Password reset successfully")
}

func (h *AuthHandler) SendOTP(c *fiber.Ctx) error {
	var req auth.SendOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := h.validator.Struct(req); err != nil {
		return h.handleValidationError(c, err)
	}

	err := h.authService.SendOTP(c.Context(), req)
	if err != nil {
		return h.handleError(c, err)
	}

	return response.Success(c, nil, "OTP sent successfully")
}

func (h *AuthHandler) VerifyOTP(c *fiber.Ctx) error {
	var req auth.VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := h.validator.Struct(req); err != nil {
		return h.handleValidationError(c, err)
	}

	err := h.authService.VerifyOTP(c.Context(), req)
	if err != nil {
		return h.handleError(c, err)
	}

	return response.Success(c, nil, "OTP verified successfully")
}

func (h *AuthHandler) UploadProfileImage(c *fiber.Ctx) error {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		return response.Unauthorized(c)
	}

	file, err := c.FormFile("image")
	if err != nil {
		return response.BadRequest(c, "No image file provided")
	}

	result, err := h.authService.UploadProfileImage(c.Context(), userID, file)
	if err != nil {
		return h.handleError(c, err)
	}

	return response.Success(c, result, "Profile image uploaded successfully")
}

func (h *AuthHandler) getUserIDFromContext(c *fiber.Ctx) (uuid.UUID, error) {
	userID, err := jwt.GetUserID(c)
	if err != nil {
		return uuid.Nil, err
	}

	return uuid.Parse(userID)
}

func (h *AuthHandler) handleError(c *fiber.Ctx, err error) error {
	switch err {
	case auth.ErrUserNotFound:
		return response.NotFound(c, "User not found")
	case auth.ErrInvalidCredentials:
		return response.Unauthorized(c, "Invalid credentials")
	case auth.ErrUserAlreadyExists:
		return response.Conflict(c, "User already exists")
	case auth.ErrPhoneAlreadyExists:
		return response.Conflict(c, "Phone number already exists")
	case auth.ErrEmailAlreadyExists:
		return response.Conflict(c, "Email already exists")
	case auth.ErrUserNotVerified:
		return response.Forbidden(c, "User not verified")
	case auth.ErrInvalidToken, auth.ErrTokenExpired:
		return response.Unauthorized(c, "Invalid or expired token")
	case auth.ErrInvalidOTP:
		return response.BadRequest(c, "Invalid OTP")
	case auth.ErrOTPExpired, auth.ErrOTPNotFound:
		return response.BadRequest(c, "OTP expired or not found")
	case auth.ErrSamePassword:
		return response.BadRequest(c, "New password cannot be the same as current password")
	case auth.ErrWeakPassword:
		return response.BadRequest(c, "Password does not meet strength requirements")
	case auth.ErrInvalidPhone:
		return response.BadRequest(c, "Invalid phone number format")
	case auth.ErrInvalidEmail:
		return response.BadRequest(c, "Invalid email format")
	case auth.ErrUserInactive:
		return response.Forbidden(c, "User account is inactive")
	case auth.ErrInvalidImageFormat:
		return response.BadRequest(c, "Invalid image format")
	case auth.ErrImageTooLarge:
		return response.BadRequest(c, "Image file too large")
	default:
		h.logger.Error("Unhandled auth error", logger.Fields{"error": err.Error()})
		return response.InternalServerError(c, "An unexpected error occurred")
	}
}

func (h *AuthHandler) handleValidationError(c *fiber.Ctx, err error) error {
	var errors []response.ErrorDetail

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, validationErr := range validationErrors {
			errors = append(errors, response.ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: h.getValidationMessage(validationErr),
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

	return response.ValidationError(c, errors)
}

func (h *AuthHandler) getValidationMessage(err validator.FieldError) string {
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
