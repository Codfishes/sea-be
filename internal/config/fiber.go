package config

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"sea-catering-backend/pkg/logger"
)

type FiberConfig struct {
	AppName      string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	BodyLimit    int
	Prefork      bool
}

func LoadFiberConfig() *FiberConfig {
	bodyLimit := 4 * 1024 * 1024
	if envBodyLimit := os.Getenv("MAX_FILE_SIZE"); envBodyLimit != "" {
		if parsed, err := strconv.Atoi(envBodyLimit); err == nil {
			bodyLimit = parsed
		}
	}

	prefork := false
	if envPrefork := os.Getenv("FIBER_PREFORK"); envPrefork != "" {
		if parsed, err := strconv.ParseBool(envPrefork); err == nil {
			prefork = parsed
		}
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	appName := os.Getenv("APP_NAME")
	if appName == "" {
		appName = "SEA Catering Backend"
	}

	return &FiberConfig{
		AppName:      appName,
		Port:         port,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
		BodyLimit:    bodyLimit,
		Prefork:      prefork,
	}
}

func NewFiber(log *logger.Logger) *fiber.App {
	config := LoadFiberConfig()

	app := fiber.New(fiber.Config{
		AppName:                   config.AppName,
		ServerHeader:              "SEA-Catering-Backend",
		StrictRouting:             false,
		CaseSensitive:             false,
		UnescapePath:              false,
		BodyLimit:                 config.BodyLimit,
		ReadTimeout:               config.ReadTimeout,
		WriteTimeout:              config.WriteTimeout,
		IdleTimeout:               config.IdleTimeout,
		DisableKeepalive:          false,
		DisableDefaultDate:        false,
		DisableDefaultContentType: false,
		DisableHeaderNormalizing:  false,
		DisableStartupMessage:     false,
		EnablePrintRoutes:         true,
		Prefork:                   config.Prefork,

		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			message := "Internal Server Error"

			if ctx.Response().StatusCode() != fiber.StatusOK {

				var fiberErr *fiber.Error
				if errors.As(err, &fiberErr) {
					code = fiberErr.Code
					message = fiberErr.Message
				}

				if code >= 400 {
					log.WithFields(logger.Fields{
						"error":      err.Error(),
						"path":       ctx.Path(),
						"method":     ctx.Method(),
						"ip":         ctx.IP(),
						"status":     code,
						"request_id": ctx.Get("X-Request-ID"),
					}).Error("Fiber error occurred")
				}

				return ctx.Status(code).JSON(fiber.Map{
					"success":   false,
					"error":     true,
					"message":   message,
					"status":    code,
					"timestamp": time.Now().UTC(),
				})
			}

			return nil
		},

		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(ctx *fiber.Ctx, e interface{}) {
			log.WithFields(logger.Fields{
				"panic":  e,
				"path":   ctx.Path(),
				"method": ctx.Method(),
				"ip":     ctx.IP(),
			}).Error("Application panic recovered")
		},
	}))

	return app
}

func GetPort() string {
	config := LoadFiberConfig()
	return config.Port
}
