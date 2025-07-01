package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"sea-catering-backend/pkg/jwt"
	"sea-catering-backend/pkg/logger"
)

type Interface interface {
	RequestID() fiber.Handler
	Logger() fiber.Handler
	CORS() fiber.Handler
	RateLimit() fiber.Handler
	AuthMiddleware() fiber.Handler
	AdminMiddleware() fiber.Handler
	OptionalAuth() fiber.Handler
	GetRequestID(c *fiber.Ctx) string
}

type middleware struct {
	logger     *logger.Logger
	jwtService jwt.Interface
}

func New(logger *logger.Logger, jwtService jwt.Interface) Interface {
	return &middleware{
		logger:     logger,
		jwtService: jwtService,
	}
}

func (m *middleware) RequestID() fiber.Handler {
	return requestid.New(requestid.Config{
		Header: "X-Request-ID",
		Generator: func() string {
			return m.generateRequestID()
		},
	})
}

func (m *middleware) Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start)
		requestID := c.Get("X-Request-ID")

		fields := logger.Fields{
			"request_id": requestID,
			"method":     c.Method(),
			"path":       c.Path(),
			"status":     c.Response().StatusCode(),
			"duration":   duration.String(),
			"ip":         c.IP(),
			"user_agent": c.Get("User-Agent"),
		}

		if userID := c.Locals("user_id"); userID != nil {
			fields["user_id"] = userID
		}

		if err != nil && c.Response().StatusCode() >= 400 {
			fields["error"] = err.Error()
			m.logger.Error("HTTP request failed", fields)
		} else if c.Response().StatusCode() >= 400 {
			m.logger.Error("HTTP request failed", fields)
		} else {
			m.logger.Info("HTTP request completed", fields)
		}

		return err
	}
}

func (m *middleware) CORS() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID",
		ExposeHeaders:    "X-Request-ID",
		AllowCredentials: false,
		MaxAge:           300,
	})
}

func (m *middleware) RateLimit() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many requests",
			})
		},
	})
}

func (m *middleware) AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Authorization header required",
			})
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid authorization header format",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Token required",
			})
		}

		claims, err := m.jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			m.logger.Error("Token validation failed", logger.Fields{
				"error": err.Error(),
				"token": tokenString[:10] + "...",
			})

			if strings.Contains(err.Error(), "token has expired") {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"success": false,
					"error":   "Token has expired",
					"code":    "TOKEN_EXPIRED",
				})
			}

			if strings.Contains(err.Error(), "token is malformed") {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"success": false,
					"error":   "Invalid token format",
					"code":    "TOKEN_MALFORMED",
				})
			}

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid or expired token",
				"code":    "TOKEN_INVALID",
			})
		}

		c.Locals("user", claims)
		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("user_role", claims.Role)

		m.logger.Debug("User authenticated successfully", logger.Fields{
			"user_id": claims.UserID,
			"email":   claims.Email,
			"role":    claims.Role,
		})

		return c.Next()
	}
}

func (m *middleware) AdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Authorization header required",
			})
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid authorization header format",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Token required",
			})
		}

		claims, err := m.jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid or expired token",
			})
		}

		if claims.Role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error":   "Admin access required",
			})
		}

		c.Locals("user", claims)
		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("user_role", claims.Role)

		return c.Next()
	}
}

func (m *middleware) OptionalAuth() fiber.Handler {
	return OptionalAuthMiddleware()
}

func (m *middleware) GetRequestID(c *fiber.Ctx) string {
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

func (m *middleware) generateRequestID() string {

	return "req_" + time.Now().Format("20060102150405") + "_" + generateRandomString(8)
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
