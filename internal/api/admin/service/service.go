package service

import (
	"context"
	"fmt"
	"time"

	"sea-catering-backend/internal/api/admin"
	"sea-catering-backend/internal/api/admin/repository"
	authRepo "sea-catering-backend/internal/api/auth/repository"
	subscriptionRepo "sea-catering-backend/internal/api/subscriptions/repository"
	testimonialRepo "sea-catering-backend/internal/api/testimonials/repository"
	"sea-catering-backend/pkg/bcrypt"
	"sea-catering-backend/pkg/jwt"
	"sea-catering-backend/pkg/logger"
)

type AdminService interface {
	AdminLogin(ctx context.Context, req admin.AdminLoginRequest) (*admin.AdminLoginResponse, error)
	GetDashboardStats(ctx context.Context) (*admin.DashboardStatsResponse, error)
	GetDashboardStatsWithFilter(ctx context.Context, startDate, endDate time.Time) (*admin.DashboardStatsResponse, error)
	ApproveTestimonial(ctx context.Context, testimonialID string) error
	RejectTestimonial(ctx context.Context, testimonialID string) error

	GetAllUsers(ctx context.Context, req admin.UserListRequest) (*admin.UserListResponse, error)
	GetUserByID(ctx context.Context, userID string) (*admin.UserResponse, error)
	UpdateUserStatus(ctx context.Context, userID string, req admin.UpdateUserStatusRequest) error
	DeleteUser(ctx context.Context, userID string) error

	SearchSubscriptions(ctx context.Context, req admin.SubscriptionSearchRequest) (*admin.SubscriptionSearchListResponse, error)
	ForceCancelSubscription(ctx context.Context, subscriptionID string, req admin.ForceCancelSubscriptionRequest) (*admin.ForceCancelSubscriptionResponse, error)
}

type adminService struct {
	adminRepo        repository.AdminRepository
	subscriptionRepo subscriptionRepo.SubscriptionRepository
	testimonialRepo  testimonialRepo.TestimonialRepository
	userRepo         authRepo.UserRepository
	jwtService       jwt.Interface
	bcryptService    bcrypt.Interface
	logger           *logger.Logger
}

func NewAdminService(
	adminRepo repository.AdminRepository,
	subscriptionRepo subscriptionRepo.SubscriptionRepository,
	testimonialRepo testimonialRepo.TestimonialRepository,
	userRepo authRepo.UserRepository,
	jwtService jwt.Interface,
	bcryptService bcrypt.Interface,
	logger *logger.Logger,
) AdminService {
	return &adminService{
		adminRepo:        adminRepo,
		subscriptionRepo: subscriptionRepo,
		testimonialRepo:  testimonialRepo,
		userRepo:         userRepo,
		jwtService:       jwtService,
		bcryptService:    bcryptService,
		logger:           logger,
	}
}

func (s *adminService) AdminLogin(ctx context.Context, req admin.AdminLoginRequest) (*admin.AdminLoginResponse, error) {

	adminUser, err := s.adminRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if err == admin.ErrAdminNotFound {
			return nil, admin.ErrInvalidCredentials
		}
		s.logger.Error("Failed to get admin by email", logger.Fields{"error": err.Error()})
		return nil, err
	}

	err = s.bcryptService.ComparePassword(adminUser.Password, req.Password)
	if err != nil {
		return nil, admin.ErrInvalidCredentials
	}

	accessToken, err := s.jwtService.GenerateAccessToken(adminUser.ID, adminUser.Email, "admin")
	if err != nil {
		s.logger.Error("Failed to generate access token", logger.Fields{"error": err.Error()})
		return nil, err
	}

	s.logger.Info("Admin logged in successfully", logger.Fields{
		"admin_id": adminUser.ID,
		"email":    adminUser.Email,
		"role":     adminUser.Role,
	})

	return &admin.AdminLoginResponse{
		AccessToken:      accessToken.AccessToken,
		TokenType:        "Bearer",
		ExpiresInMinutes: 1440,
		Admin: admin.AdminResponse{
			ID:        adminUser.ID,
			Email:     adminUser.Email,
			Name:      adminUser.Name,
			Role:      adminUser.Role,
			CreatedAt: adminUser.CreatedAt,
		},
	}, nil
}

func (s *adminService) GetDashboardStats(ctx context.Context) (*admin.DashboardStatsResponse, error) {

	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)

	s.logger.Info("Getting dashboard stats with default period", logger.Fields{
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	})

	return s.GetDashboardStatsWithFilter(ctx, startDate, endDate)
}

func (s *adminService) GetDashboardStatsWithFilter(ctx context.Context, startDate, endDate time.Time) (*admin.DashboardStatsResponse, error) {
	s.logger.Info("Getting dashboard stats with filter", logger.Fields{
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	})

	stats := &admin.DashboardStatsResponse{}

	subscriptionStats, err := s.subscriptionRepo.GetSubscriptionStats(ctx, startDate, endDate)
	if err != nil {
		s.logger.Error("Failed to get subscription stats", logger.Fields{"error": err.Error()})
		return nil, err
	}

	stats.TotalSubscriptions = subscriptionStats.TotalSubscriptions
	stats.ActiveSubscriptions = subscriptionStats.ActiveSubscriptions
	stats.NewSubscriptions = subscriptionStats.NewSubscriptions
	stats.MonthlyRevenue = subscriptionStats.MonthlyRevenue
	stats.CancelledSubscriptions = subscriptionStats.CancelledSubscriptions
	stats.Reactivations = subscriptionStats.Reactivations

	if stats.ActiveSubscriptions > 0 {
		growthBase := float64(stats.NewSubscriptions + stats.Reactivations)
		stats.SubscriptionGrowth = (growthBase / float64(stats.ActiveSubscriptions)) * 100

		stats.SubscriptionGrowth = float64(int(stats.SubscriptionGrowth*100)) / 100
	} else {
		stats.SubscriptionGrowth = 0.0
	}

	s.logger.Debug("Subscription growth calculation", logger.Fields{
		"new_subscriptions":    stats.NewSubscriptions,
		"reactivations":        stats.Reactivations,
		"active_subscriptions": stats.ActiveSubscriptions,
		"growth_percentage":    stats.SubscriptionGrowth,
	})

	prevStartDate := startDate.AddDate(0, -1, 0)
	prevEndDate := startDate.Add(-time.Nanosecond)

	prevStats, err := s.subscriptionRepo.GetSubscriptionStats(ctx, prevStartDate, prevEndDate)
	if err != nil {
		s.logger.Warn("Failed to get previous period stats for revenue growth", logger.Fields{
			"error": err.Error(),
		})
		stats.RevenueGrowth = 0.0
	} else {
		if prevStats.MonthlyRevenue > 0 {
			revenueChange := stats.MonthlyRevenue - prevStats.MonthlyRevenue
			stats.RevenueGrowth = (revenueChange / prevStats.MonthlyRevenue) * 100

			stats.RevenueGrowth = float64(int(stats.RevenueGrowth*100)) / 100
		} else if stats.MonthlyRevenue > 0 {

			stats.RevenueGrowth = 100.0
		} else {
			stats.RevenueGrowth = 0.0
		}
	}

	s.logger.Debug("Revenue growth calculation", logger.Fields{
		"current_revenue":  stats.MonthlyRevenue,
		"previous_revenue": prevStats.MonthlyRevenue,
		"revenue_growth":   stats.RevenueGrowth,
	})

	pendingTestimonials, err := s.testimonialRepo.CountPending(ctx)
	if err != nil {
		s.logger.Error("Failed to get pending testimonials count", logger.Fields{"error": err.Error()})
		pendingTestimonials = 0
	}
	stats.PendingTestimonials = pendingTestimonials

	totalUsers, err := s.userRepo.CountActiveUsers(ctx)
	if err != nil {
		s.logger.Error("Failed to get total users count", logger.Fields{"error": err.Error()})
		totalUsers = 0
	}
	stats.TotalUsers = totalUsers

	s.logger.Info("Dashboard stats retrieved successfully", logger.Fields{
		"period_start":            startDate.Format("2006-01-02"),
		"period_end":              endDate.Format("2006-01-02"),
		"total_subscriptions":     stats.TotalSubscriptions,
		"active_subscriptions":    stats.ActiveSubscriptions,
		"new_subscriptions":       stats.NewSubscriptions,
		"reactivations":           stats.Reactivations,
		"subscription_growth":     stats.SubscriptionGrowth,
		"monthly_revenue":         stats.MonthlyRevenue,
		"revenue_growth":          stats.RevenueGrowth,
		"total_users":             stats.TotalUsers,
		"cancelled_subscriptions": stats.CancelledSubscriptions,
		"pending_testimonials":    stats.PendingTestimonials,
	})

	return stats, nil
}

func (s *adminService) ApproveTestimonial(ctx context.Context, testimonialID string) error {

	testimonial, err := s.testimonialRepo.GetByID(ctx, testimonialID)
	if err != nil {
		return err
	}

	if testimonial.IsApproved {
		return nil
	}

	testimonial.IsApproved = true
	testimonial.UpdatedAt = time.Now()

	if err := s.testimonialRepo.Update(ctx, testimonial); err != nil {
		s.logger.Error("Failed to approve testimonial", logger.Fields{
			"error":          err.Error(),
			"testimonial_id": testimonialID,
		})
		return err
	}

	s.logger.Info("Testimonial approved successfully", logger.Fields{
		"testimonial_id": testimonialID,
		"customer_name":  testimonial.CustomerName,
	})

	return nil
}

func (s *adminService) RejectTestimonial(ctx context.Context, testimonialID string) error {

	testimonial, err := s.testimonialRepo.GetByID(ctx, testimonialID)
	if err != nil {
		return err
	}

	if !testimonial.IsApproved {
		return nil
	}

	testimonial.IsApproved = false
	testimonial.UpdatedAt = time.Now()

	if err := s.testimonialRepo.Update(ctx, testimonial); err != nil {
		s.logger.Error("Failed to reject testimonial", logger.Fields{
			"error":          err.Error(),
			"testimonial_id": testimonialID,
		})
		return err
	}

	s.logger.Info("Testimonial rejected successfully", logger.Fields{
		"testimonial_id": testimonialID,
		"customer_name":  testimonial.CustomerName,
	})

	return nil
}

func (s *adminService) GetAllUsers(ctx context.Context, req admin.UserListRequest) (*admin.UserListResponse, error) {
	users, meta, err := s.adminRepo.GetAllUsers(ctx, req)
	if err != nil {
		s.logger.Error("Failed to get all users", logger.Fields{
			"error": err.Error(),
			"req":   req,
		})
		return nil, err
	}

	s.logger.Info("Retrieved users list", logger.Fields{
		"count": len(users),
		"page":  req.Page,
		"limit": req.Limit,
	})

	return &admin.UserListResponse{
		Users: users,
		Meta:  meta,
	}, nil
}

func (s *adminService) GetUserByID(ctx context.Context, userID string) (*admin.UserResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	user, err := s.adminRepo.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user by ID", logger.Fields{
			"error":   err.Error(),
			"user_id": userID,
		})
		return nil, err
	}

	s.logger.Info("Retrieved user details", logger.Fields{
		"user_id":   userID,
		"user_name": user.Name,
	})

	return user, nil
}

func (s *adminService) UpdateUserStatus(ctx context.Context, userID string, req admin.UpdateUserStatusRequest) error {
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	user, err := s.adminRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	err = s.adminRepo.UpdateUserStatus(ctx, userID, req.IsActive, req.Reason)
	if err != nil {
		s.logger.Error("Failed to update user status", logger.Fields{
			"error":     err.Error(),
			"user_id":   userID,
			"is_active": req.IsActive,
			"reason":    req.Reason,
		})
		return err
	}

	action := "activated"
	if !req.IsActive {
		action = "deactivated"
	}

	s.logger.Info("User status updated", logger.Fields{
		"user_id":   userID,
		"user_name": user.Name,
		"action":    action,
		"reason":    req.Reason,
	})

	return nil
}

func (s *adminService) DeleteUser(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	user, err := s.adminRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user.SubscriptionCount > 0 {
		return fmt.Errorf("cannot delete user with %d active subscriptions", user.SubscriptionCount)
	}

	err = s.adminRepo.DeleteUser(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to delete user", logger.Fields{
			"error":   err.Error(),
			"user_id": userID,
		})
		return err
	}

	s.logger.Info("User deleted", logger.Fields{
		"user_id":   userID,
		"user_name": user.Name,
		"user_email": func() string {
			if user.Email != nil {
				return *user.Email
			}
			return ""
		}(),
	})

	return nil
}

func (s *adminService) SearchSubscriptions(ctx context.Context, req admin.SubscriptionSearchRequest) (*admin.SubscriptionSearchListResponse, error) {

	if req.DateFrom != "" && req.DateTo != "" {
		dateFrom, err := time.Parse("2006-01-02", req.DateFrom)
		if err != nil {
			return nil, fmt.Errorf("invalid date_from format: %w", err)
		}

		dateTo, err := time.Parse("2006-01-02", req.DateTo)
		if err != nil {
			return nil, fmt.Errorf("invalid date_to format: %w", err)
		}

		if dateFrom.After(dateTo) {
			return nil, fmt.Errorf("date_from cannot be after date_to")
		}
	}

	if req.MinPrice > 0 && req.MaxPrice > 0 && req.MinPrice > req.MaxPrice {
		return nil, fmt.Errorf("min_price cannot be greater than max_price")
	}

	subscriptions, meta, err := s.adminRepo.SearchSubscriptions(ctx, req)
	if err != nil {
		s.logger.Error("Failed to search subscriptions", logger.Fields{
			"error": err.Error(),
			"req":   req,
		})
		return nil, err
	}

	s.logger.Info("Retrieved subscriptions search results", logger.Fields{
		"count":  len(subscriptions),
		"page":   req.Page,
		"limit":  req.Limit,
		"search": req.Search,
		"status": req.Status,
	})

	return &admin.SubscriptionSearchListResponse{
		Subscriptions: subscriptions,
		Meta:          meta,
	}, nil
}

func (s *adminService) ForceCancelSubscription(ctx context.Context, subscriptionID string, req admin.ForceCancelSubscriptionRequest) (*admin.ForceCancelSubscriptionResponse, error) {
	if subscriptionID == "" {
		return nil, fmt.Errorf("subscription ID is required")
	}

	subscription, err := s.adminRepo.GetSubscriptionForCancel(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	if subscription.Status == "cancelled" {
		return nil, fmt.Errorf("subscription is already cancelled")
	}

	err = s.adminRepo.ForceCancelSubscription(ctx, subscriptionID, req.Reason, req.AdminComments)
	if err != nil {
		s.logger.Error("Failed to force cancel subscription", logger.Fields{
			"error":           err.Error(),
			"subscription_id": subscriptionID,
			"reason":          req.Reason,
		})
		return nil, err
	}

	cancelledAt := time.Now()
	var refundID *string

	if req.RefundAmount != nil && *req.RefundAmount > 0 {

		s.logger.Info("Refund requested for cancelled subscription", logger.Fields{
			"subscription_id": subscriptionID,
			"refund_amount":   *req.RefundAmount,
			"reason":          req.Reason,
		})
	}

	if req.NotifyUser {
		s.logger.Info("User notification sent for subscription cancellation", logger.Fields{
			"subscription_id": subscriptionID,
			"user_id":         subscription.UserID,
			"user_name":       subscription.UserName,
		})
	}

	s.logger.Info("Subscription force cancelled by admin", logger.Fields{
		"subscription_id": subscriptionID,
		"user_id":         subscription.UserID,
		"user_name":       subscription.UserName,
		"reason":          req.Reason,
		"refund_amount":   req.RefundAmount,
		"admin_comments":  req.AdminComments,
	})

	response := &admin.ForceCancelSubscriptionResponse{
		SubscriptionID: subscriptionID,
		CancelledAt:    cancelledAt,
		Reason:         req.Reason,
		RefundAmount:   req.RefundAmount,
		RefundID:       refundID,
		Message:        "Subscription has been successfully cancelled by admin",
	}

	return response, nil
}
