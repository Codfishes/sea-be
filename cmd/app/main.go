package main

import (
	"github.com/gofiber/fiber/v2"
	"log"
	"os"

	"github.com/joho/godotenv"

	"sea-catering-backend/database/postgres"

	authHandler "sea-catering-backend/internal/api/auth/handler"
	authRepository "sea-catering-backend/internal/api/auth/repository"
	authService "sea-catering-backend/internal/api/auth/service"

	mealPlansHandler "sea-catering-backend/internal/api/meal_plans/handler"
	mealPlansRepository "sea-catering-backend/internal/api/meal_plans/repository"
	mealPlansService "sea-catering-backend/internal/api/meal_plans/service"

	subscriptionsHandler "sea-catering-backend/internal/api/subscriptions/handler"
	subscriptionsRepository "sea-catering-backend/internal/api/subscriptions/repository"
	subscriptionsService "sea-catering-backend/internal/api/subscriptions/service"

	testimonialsHandler "sea-catering-backend/internal/api/testimonials/handler"
	testimonialsRepository "sea-catering-backend/internal/api/testimonials/repository"
	testimonialsService "sea-catering-backend/internal/api/testimonials/service"

	adminHandler "sea-catering-backend/internal/api/admin/handler"
	adminRepository "sea-catering-backend/internal/api/admin/repository"
	adminService "sea-catering-backend/internal/api/admin/service"

	"sea-catering-backend/internal/config"
	"sea-catering-backend/internal/middleware"
	"sea-catering-backend/pkg/bcrypt"
	"sea-catering-backend/pkg/email"
	"sea-catering-backend/pkg/jwt"
	"sea-catering-backend/pkg/logger"
	"sea-catering-backend/pkg/redis"
	"sea-catering-backend/pkg/s3"
	"sea-catering-backend/pkg/utils"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	appLogger := config.NewLogger()
	appLogger.Info("Starting SEA Catering Backend", logger.Fields{
		"version":     "1.0.0",
		"environment": os.Getenv("APP_ENV"),
	})

	appLogger.Info("Connecting to database...")
	db, err := postgres.New()
	if err != nil {
		appLogger.Fatal("Failed to connect to database", logger.Fields{
			"error": err.Error(),
		})
	}
	defer postgres.Close(db)
	appLogger.Info("Database connected successfully")

	appLogger.Info("Connecting to Redis...")
	redisClient, err := redis.New()
	if err != nil {
		appLogger.Fatal("Failed to connect to Redis", logger.Fields{
			"error": err.Error(),
		})
	}
	defer redisClient.Close()
	appLogger.Info("Redis connected successfully")

	appLogger.Info("Initializing services...")

	validator := config.NewValidator()
	jwtService := jwt.New()
	bcryptService := bcrypt.New()
	utilsService := utils.New()

	emailService := email.New()
	s3Service, err := s3.New()
	if err != nil {
		appLogger.Fatal("Failed to initialize S3 service", logger.Fields{
			"error": err.Error(),
		})
	}

	fiberApp := config.NewFiber(appLogger)

	middlewareService := middleware.New(appLogger, jwtService)

	fiberApp.Use(middlewareService.RequestID())
	fiberApp.Use(middlewareService.Logger())
	fiberApp.Use(middlewareService.CORS())
	fiberApp.Use(middlewareService.RateLimit())

	appLogger.Info("Services initialized successfully")

	userRepo := authRepository.NewUserRepository(db)
	mealPlanRepo := mealPlansRepository.NewMealPlanRepository(db)
	subscriptionRepo := subscriptionsRepository.NewSubscriptionRepository(db, appLogger, utilsService)
	testimonialRepo := testimonialsRepository.NewTestimonialRepository(db)
	adminRepo := adminRepository.NewAdminRepository(db)

	authSvc := authService.NewAuthService(
		userRepo,
		jwtService,
		bcryptService,
		redisClient,
		emailService,
		s3Service,
		utilsService,
		appLogger,
	)

	mealPlanSvc := mealPlansService.NewMealPlanService(
		mealPlanRepo,
		utilsService,
		appLogger,
	)

	subscriptionSvc := subscriptionsService.NewSubscriptionService(
		subscriptionRepo,
		mealPlanRepo,
		utilsService,
		appLogger,
	)

	testimonialSvc := testimonialsService.NewTestimonialService(
		testimonialRepo,
		utilsService,
		appLogger,
	)

	adminSvc := adminService.NewAdminService(
		adminRepo,
		subscriptionRepo,
		testimonialRepo,
		userRepo,
		jwtService,
		bcryptService,
		appLogger,
	)

	authHdlr := authHandler.NewAuthHandler(authSvc, validator, appLogger)
	mealPlanHdlr := mealPlansHandler.NewMealPlanHandler(mealPlanSvc, validator, middlewareService, appLogger)
	subscriptionHdlr := subscriptionsHandler.NewSubscriptionHandler(subscriptionSvc, validator, middlewareService, appLogger)
	testimonialHdlr := testimonialsHandler.NewTestimonialHandler(testimonialSvc, validator, middlewareService, appLogger)
	adminHdlr := adminHandler.NewAdminHandler(adminSvc, validator, middlewareService, appLogger)

	api := fiberApp.Group("/api/v1")

	authHdlr.RegisterRoutes(api)

	mealPlanHdlr.RegisterRoutes(api)

	subscriptionHdlr.RegisterRoutes(api)

	testimonialHdlr.RegisterRoutes(api)

	adminHdlr.RegisterRoutes(api)

	fiberApp.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "healthy",
			"service":   "sea-catering-backend",
			"version":   "1.0.0",
			"timestamp": "2024-01-01T00:00:00Z",
		})
	})

	fiberApp.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Welcome to SEA Catering Backend API",
			"version": "1.0.0",
			"docs":    "/api/v1/docs",
			"endpoints": fiber.Map{
				"auth": fiber.Map{
					"register":        "POST /api/v1/auth/register",
					"login":           "POST /api/v1/auth/login",
					"logout":          "POST /api/v1/auth/logout",
					"profile":         "GET /api/v1/user/profile",
					"update_profile":  "PUT /api/v1/user/profile",
					"change_password": "POST /api/v1/user/change-password",
					"upload_image":    "POST /api/v1/user/profile/image",
				},
				"meal_plans": fiber.Map{
					"list":        "GET /api/v1/meal-plans",
					"active":      "GET /api/v1/meal-plans/active",
					"search":      "GET /api/v1/meal-plans/search?q={query}",
					"get_by_id":   "GET /api/v1/meal-plans/{id}",
					"popular":     "GET /api/v1/meal-plans/popular",
					"create":      "POST /api/v1/meal-plans/admin (Admin only)",
					"update":      "PUT /api/v1/meal-plans/admin/{id} (Admin only)",
					"delete":      "DELETE /api/v1/meal-plans/admin/{id} (Admin only)",
					"activate":    "PATCH /api/v1/meal-plans/admin/{id}/activate (Admin only)",
					"deactivate":  "PATCH /api/v1/meal-plans/admin/{id}/deactivate (Admin only)",
					"bulk_status": "PATCH /api/v1/meal-plans/admin/bulk-status (Admin only)",
					"stats":       "GET /api/v1/meal-plans/admin/stats (Admin only)",
				},
				"subscriptions": fiber.Map{
					"create":     "POST /api/v1/subscriptions (Auth required)",
					"my":         "GET /api/v1/subscriptions/my (Auth required)",
					"get_by_id":  "GET /api/v1/subscriptions/{id} (Auth required)",
					"update":     "PUT /api/v1/subscriptions/{id} (Auth required)",
					"pause":      "PUT /api/v1/subscriptions/{id}/pause (Auth required)",
					"resume":     "PUT /api/v1/subscriptions/{id}/resume (Auth required)",
					"reactivate": "PUT /api/v1/subscriptions/{id}/reactivate (Auth required)",
					"cancel":     "DELETE /api/v1/subscriptions/{id} (Auth required)",
					"stats":      "GET /api/v1/subscriptions/admin/stats (Admin only)",
				},
				"testimonials": fiber.Map{
					"create":        "POST /api/v1/testimonials",
					"get_approved":  "GET /api/v1/testimonials",
					"admin_get_all": "GET /api/v1/testimonials/admin/all (Admin only)",
					"admin_approve": "PUT /api/v1/testimonials/admin/{id}/approve (Admin only)",
					"admin_reject":  "PUT /api/v1/testimonials/admin/{id}/reject (Admin only)",
					"admin_delete":  "DELETE /api/v1/testimonials/admin/{id} (Admin only)",
				},
				"admin": fiber.Map{
					"login":               "POST /api/v1/admin/login",
					"dashboard":           "GET /api/v1/admin/dashboard (Admin only)",
					"dashboard_filter":    "POST /api/v1/admin/dashboard/filter (Admin only)",
					"approve_testimonial": "PUT /api/v1/admin/testimonials/{id}/approve (Admin only)",
					"reject_testimonial":  "PUT /api/v1/admin/testimonials/{id}/reject (Admin only)",
				},
			},
			"business_info": fiber.Map{
				"name":    "SEA Catering",
				"slogan":  "Healthy Meals, Anytime, Anywhere",
				"manager": "Brian",
				"phone":   "08123456789",
				"features": []string{
					"Customizable healthy meal plans",
					"Delivery across Indonesia",
					"Detailed nutritional information",
					"Flexible subscription management",
					"Professional meal planning",
				},
			},
		})
	})

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	appLogger.Info("Server starting...", logger.Fields{
		"port": port,
	})

	if err := fiberApp.Listen(":" + port); err != nil {
		appLogger.Fatal("Server failed to start", logger.Fields{
			"error": err.Error(),
		})
	}
}
