package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type CORSConfig struct {
	AllowOrigins     string
	AllowMethods     string
	AllowHeaders     string
	ExposeHeaders    string
	AllowCredentials bool
	MaxAge           int
}

func LoadCORSConfig() *CORSConfig {
	allowOrigins := os.Getenv("CORS_ALLOW_ORIGINS")
	if allowOrigins == "" {

		allowOrigins = "http://localhost:3000,http://localhost:3001,http://127.0.0.1:3000"
	}

	allowMethods := os.Getenv("CORS_ALLOW_METHODS")
	if allowMethods == "" {
		allowMethods = "GET,POST,PUT,DELETE,PATCH,OPTIONS"
	}

	allowHeaders := os.Getenv("CORS_ALLOW_HEADERS")
	if allowHeaders == "" {
		allowHeaders = "Origin,Content-Type,Accept,Authorization,X-Request-ID,X-Requested-With"
	}

	exposeHeaders := os.Getenv("CORS_EXPOSE_HEADERS")
	if exposeHeaders == "" {
		exposeHeaders = "X-Request-ID,X-Total-Count"
	}

	allowCredentials := os.Getenv("CORS_ALLOW_CREDENTIALS") == "true"

	return &CORSConfig{
		AllowOrigins:     allowOrigins,
		AllowMethods:     allowMethods,
		AllowHeaders:     allowHeaders,
		ExposeHeaders:    exposeHeaders,
		AllowCredentials: allowCredentials,
		MaxAge:           300,
	}
}

func NewCORSMiddleware() fiber.Handler {
	config := LoadCORSConfig()

	return cors.New(cors.Config{
		AllowOrigins:     config.AllowOrigins,
		AllowMethods:     config.AllowMethods,
		AllowHeaders:     config.AllowHeaders,
		ExposeHeaders:    config.ExposeHeaders,
		AllowCredentials: config.AllowCredentials,
		MaxAge:           config.MaxAge,
		AllowOriginsFunc: func(origin string) bool {

			if os.Getenv("APP_ENV") == "development" {
				return true
			}

			allowedOrigins := strings.Split(config.AllowOrigins, ",")
			for _, allowedOrigin := range allowedOrigins {
				if strings.TrimSpace(allowedOrigin) == origin {
					return true
				}
			}
			return false
		},
	})
}
