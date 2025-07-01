package config

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"sea-catering-backend/internal/middleware"
	"sea-catering-backend/pkg/bcrypt"
	"sea-catering-backend/pkg/email"
	"sea-catering-backend/pkg/jwt"
	"sea-catering-backend/pkg/logger"
	"sea-catering-backend/pkg/midtrans"
	"sea-catering-backend/pkg/s3"
	"sea-catering-backend/pkg/utils"
)

type ServerOption func(*Server) error

type Server struct {
	app        *fiber.App
	db         *sqlx.DB
	redis      *redis.Client
	logger     *logger.Logger
	validator  *validator.Validate
	jwt        jwt.Interface
	bcrypt     bcrypt.Interface
	email      email.Interface
	s3         s3.Interface
	midtrans   midtrans.Interface
	utils      utils.Interface
	middleware middleware.Interface
	handlers   []Handler
	config     *ServerConfig
}

type Handler interface {
	RegisterRoutes(router fiber.Router)
}

type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	Environment     string
}

func LoadServerConfig() *ServerConfig {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	return &ServerConfig{
		Port:            port,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		ShutdownTimeout: 30 * time.Second,
		Environment:     env,
	}
}

func NewServer(options ...ServerOption) (*Server, error) {
	config := LoadServerConfig()

	server := &Server{
		config:   config,
		handlers: make([]Handler, 0),
	}

	for _, option := range options {
		if err := option(server); err != nil {
			return nil, fmt.Errorf("failed to apply server option: %w", err)
		}
	}

	if server.logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if server.app == nil {
		return nil, fmt.Errorf("fiber app is required")
	}
	if server.db == nil {
		return nil, fmt.Errorf("database connection is required")
	}

	server.app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		ExposeHeaders:    "Content-Length,Content-Type",
		AllowCredentials: true,
	}))
	return server, nil
}

func WithFiber(app *fiber.App) ServerOption {
	return func(s *Server) error {
		s.app = app
		return nil
	}
}

func WithDatabase(db *sqlx.DB) ServerOption {
	return func(s *Server) error {
		s.db = db
		return nil
	}
}

func WithRedis(client *redis.Client) ServerOption {
	return func(s *Server) error {
		s.redis = client
		return nil
	}
}

func WithLogger(log *logger.Logger) ServerOption {
	return func(s *Server) error {
		s.logger = log
		return nil
	}
}

func WithValidator(v *validator.Validate) ServerOption {
	return func(s *Server) error {
		s.validator = v
		return nil
	}
}

func WithJWT(j jwt.Interface) ServerOption {
	return func(s *Server) error {
		s.jwt = j
		return nil
	}
}

func WithBcrypt(b bcrypt.Interface) ServerOption {
	return func(s *Server) error {
		s.bcrypt = b
		return nil
	}
}

func WithEmail(e email.Interface) ServerOption {
	return func(s *Server) error {
		s.email = e
		return nil
	}
}

func WithS3(s3Service s3.Interface) ServerOption {
	return func(s *Server) error {
		s.s3 = s3Service
		return nil
	}
}

func WithMidtrans(m midtrans.Interface) ServerOption {
	return func(s *Server) error {
		s.midtrans = m
		return nil
	}
}

func WithUtils(u utils.Interface) ServerOption {
	return func(s *Server) error {
		s.utils = u
		return nil
	}
}

func WithMiddleware(m middleware.Interface) ServerOption {
	return func(s *Server) error {
		s.middleware = m
		return nil
	}
}

func WithHandlers(handlers ...Handler) ServerOption {
	return func(s *Server) error {
		s.handlers = append(s.handlers, handlers...)
		return nil
	}
}

func (s *Server) RegisterRoutes() {

	s.app.Get("/health", s.healthCheck)
	s.app.Get("/", s.welcome)

	api := s.app.Group("/api/v1")

	for _, handler := range s.handlers {
		handler.RegisterRoutes(api)
	}

	s.logger.Info("All routes registered successfully")
}

func (s *Server) Start() error {

	s.RegisterRoutes()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		s.logger.Info("Starting server", logger.Fields{
			"port":        s.config.Port,
			"environment": s.config.Environment,
		})

		if err := s.app.Listen(":" + s.config.Port); err != nil {
			s.logger.Fatal("Failed to start server", logger.Fields{
				"error": err.Error(),
			})
		}
	}()

	s.logger.Info("Server started successfully", logger.Fields{
		"port": s.config.Port,
	})

	<-quit
	s.logger.Info("Shutting down server...")

	return s.Shutdown()
}

func (s *Server) Shutdown() error {

	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	if err := s.app.ShutdownWithContext(ctx); err != nil {
		s.logger.Error("Error during server shutdown", logger.Fields{
			"error": err.Error(),
		})
		return err
	}

	if s.db != nil {
		if err := s.db.Close(); err != nil {
			s.logger.Error("Error closing database connection", logger.Fields{
				"error": err.Error(),
			})
		}
	}

	if s.redis != nil {
		if err := s.redis.Close(); err != nil {
			s.logger.Error("Error closing Redis connection", logger.Fields{
				"error": err.Error(),
			})
		}
	}

	s.logger.Info("Server shutdown completed")
	return nil
}

func (s *Server) GetApp() *fiber.App {
	return s.app
}

func (s *Server) GetDB() *sqlx.DB {
	return s.db
}

func (s *Server) GetRedis() *redis.Client {
	return s.redis
}

func (s *Server) GetLogger() *logger.Logger {
	return s.logger
}

func (s *Server) GetValidator() *validator.Validate {
	return s.validator
}

func (s *Server) GetJWT() jwt.Interface {
	return s.jwt
}

func (s *Server) GetBcrypt() bcrypt.Interface {
	return s.bcrypt
}

func (s *Server) GetEmail() email.Interface {
	return s.email
}

func (s *Server) GetS3() s3.Interface {
	return s.s3
}

func (s *Server) GetMidtrans() midtrans.Interface {
	return s.midtrans
}

func (s *Server) GetUtils() utils.Interface {
	return s.utils
}

func (s *Server) GetMiddleware() middleware.Interface {
	return s.middleware
}

func (s *Server) healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "sea-catering-backend",
		"version":   "1.0.0",
	})
}

func (s *Server) welcome(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Welcome to SEA Catering Backend API",
		"version": "1.0.0",
		"docs":    "/api/v1/docs",
	})
}
