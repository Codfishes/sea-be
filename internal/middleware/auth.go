package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"sea-catering-backend/pkg/jwt"
	"sea-catering-backend/pkg/logger"
	"sea-catering-backend/pkg/response"
)

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Unauthorized(c, "Authorization header required")
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			return response.Unauthorized(c, "Invalid authorization header format")
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			return response.Unauthorized(c, "Token required")
		}

		jwtService := jwt.New()

		claims, err := jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			return response.Unauthorized(c, "Invalid or expired token")
		}

		c.Locals("user", claims)
		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("user_role", claims.Role)

		return c.Next()
	}
}

func OptionalAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Next()
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			return c.Next()
		}

		jwtService := jwt.New()
		claims, err := jwtService.ValidateAccessToken(tokenString)
		if err != nil {

			logger.GetInstance().Warn("Invalid token in optional auth", logger.Fields{
				"error": err.Error(),
			})
			return c.Next()
		}

		c.Locals("user", claims)
		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("user_role", claims.Role)

		return c.Next()
	}
}

func EmailVerificationMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {

		claims, ok := c.Locals("user").(*jwt.Claims)
		if !ok {
			return response.Unauthorized(c, "Authentication required")
		}

		if !claims.IsVerified {
			return response.Forbidden(c, "Email verification required")
		}

		return c.Next()
	}
}
