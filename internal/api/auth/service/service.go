package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/google/uuid"

	"sea-catering-backend/internal/api/auth"
	"sea-catering-backend/internal/api/auth/repository"
	"sea-catering-backend/internal/entity"
	"sea-catering-backend/pkg/bcrypt"
	"sea-catering-backend/pkg/email"
	"sea-catering-backend/pkg/jwt"
	"sea-catering-backend/pkg/logger"
	"sea-catering-backend/pkg/redis"
	"sea-catering-backend/pkg/s3"
	"sea-catering-backend/pkg/utils"
)

type AuthService interface {
	Register(ctx context.Context, req auth.RegisterRequest) (*auth.LoginResponse, error)
	Login(ctx context.Context, req auth.LoginRequest) (*auth.LoginResponse, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*entity.UserResponse, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req auth.UpdateProfileRequest) (*entity.UserResponse, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, req auth.ChangePasswordRequest) error
	ForgotPassword(ctx context.Context, req auth.ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, req auth.ResetPasswordRequest) error
	SendOTP(ctx context.Context, req auth.SendOTPRequest) error
	VerifyOTP(ctx context.Context, req auth.VerifyOTPRequest) error
	UploadProfileImage(ctx context.Context, userID uuid.UUID, file *multipart.FileHeader) (*auth.UploadProfileImageResponse, error)
}

type authService struct {
	userRepo      repository.UserRepository
	jwtService    jwt.Interface
	bcryptService bcrypt.Interface
	redisService  redis.Interface
	emailService  email.Interface
	s3Service     s3.Interface
	utilsService  utils.Interface
	logger        *logger.Logger
}

func NewAuthService(
	userRepo repository.UserRepository,
	jwtService jwt.Interface,
	bcryptService bcrypt.Interface,
	redisService redis.Interface,
	emailService email.Interface,
	s3Service s3.Interface,
	utilsService utils.Interface,
	logger *logger.Logger,
) AuthService {
	return &authService{
		userRepo:      userRepo,
		jwtService:    jwtService,
		bcryptService: bcryptService,
		redisService:  redisService,
		emailService:  emailService,
		s3Service:     s3Service,
		utilsService:  utilsService,
		logger:        logger,
	}
}

func (s *authService) Register(ctx context.Context, req auth.RegisterRequest) (*auth.LoginResponse, error) {

	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Error("Failed to check email existence", logger.Fields{"error": err.Error()})
		return nil, err
	}
	if exists {
		return nil, auth.ErrEmailAlreadyExists
	}

	if req.Phone != "" {
		exists, err := s.userRepo.ExistsByPhone(ctx, req.Phone)
		if err != nil {
			s.logger.Error("Failed to check phone existence", logger.Fields{"error": err.Error()})
			return nil, err
		}
		if exists {
			return nil, auth.ErrPhoneAlreadyExists
		}
	}

	hashedPassword, err := s.bcryptService.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("Failed to hash password", logger.Fields{"error": err.Error()})
		return nil, err
	}

	user := &entity.User{
		ID:        uuid.New(),
		Name:      req.Name,
		Email:     req.Email,
		Password:  hashedPassword,
		Role:      entity.RoleUser,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if req.Phone != "" {
		user.Phone = &req.Phone
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		s.logger.Error("Failed to create user", logger.Fields{"error": err.Error()})
		return nil, err
	}

	err = s.emailService.SendWelcomeEmail(user.Email, user.Name)
	if err != nil {
		s.logger.Warn("Failed to send welcome email", logger.Fields{
			"error":   err.Error(),
			"email":   user.Email,
			"user_id": user.ID.String(),
		})

	}

	err = s.SendOTP(ctx, auth.SendOTPRequest{
		Email: user.Email,
		Type:  "email_verification",
	})
	if err != nil {
		s.logger.Warn("Failed to send verification OTP", logger.Fields{
			"error": err.Error(),
			"email": user.Email,
		})

	}

	s.logger.Info("User registered successfully", logger.Fields{
		"user_id": user.ID.String(),
		"email":   user.Email,
	})

	tokenPair, err := s.jwtService.GenerateAccessToken(user.ID.String(), user.Email, user.Role)
	if err != nil {
		s.logger.Error("Failed to generate tokens", logger.Fields{"error": err.Error()})
		return nil, err
	}

	return &auth.LoginResponse{
		AccessToken: tokenPair.AccessToken,
		ExpiresAt:   tokenPair.ExpiresAt,
		User: auth.UserInfo{
			ID:              user.ID,
			Name:            user.Name,
			Email:           user.Email,
			Phone:           user.Phone,
			IsVerified:      user.IsVerified,
			ProfileImageURL: user.ProfileImageURL,
			Role:            user.Role,
		},
	}, nil
}

func (s *authService) Login(ctx context.Context, req auth.LoginRequest) (*auth.LoginResponse, error) {

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if err == auth.ErrUserNotFound {
			return nil, auth.ErrInvalidCredentials
		}
		s.logger.Error("Failed to get user by email", logger.Fields{"error": err.Error()})
		return nil, err
	}

	if !user.IsActive {
		return nil, auth.ErrUserInactive
	}

	err = s.bcryptService.ComparePassword(user.Password, req.Password)
	if err != nil {
		return nil, auth.ErrInvalidCredentials
	}

	err = s.userRepo.UpdateLastLogin(ctx, user.ID)
	if err != nil {
		s.logger.Error("Failed to update last login", logger.Fields{"error": err.Error()})

	}

	s.logger.Info("User logged in successfully", logger.Fields{
		"user_id": user.ID.String(),
		"email":   user.Email,
	})

	tokenPair, err := s.jwtService.GenerateAccessToken(user.ID.String(), user.Email, user.Role)
	if err != nil {
		s.logger.Error("Failed to generate tokens", logger.Fields{"error": err.Error()})
		return nil, err
	}

	return &auth.LoginResponse{
		AccessToken: tokenPair.AccessToken,
		ExpiresAt:   tokenPair.ExpiresAt,
		User: auth.UserInfo{
			ID:              user.ID,
			Name:            user.Name,
			Email:           user.Email,
			Phone:           user.Phone,
			IsVerified:      user.IsVerified,
			ProfileImageURL: user.ProfileImageURL,
			Role:            user.Role,
		},
	}, nil
}

func (s *authService) GetProfile(ctx context.Context, userID uuid.UUID) (*entity.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

func (s *authService) UpdateProfile(ctx context.Context, userID uuid.UUID, req auth.UpdateProfileRequest) (*entity.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		user.Name = req.Name
	}

	if req.Phone != nil {

		if user.Phone == nil || *user.Phone != *req.Phone {
			exists, err := s.userRepo.ExistsByPhone(ctx, *req.Phone)
			if err != nil {
				return nil, err
			}
			if exists {
				return nil, auth.ErrPhoneAlreadyExists
			}
		}
		user.Phone = req.Phone

		user.PhoneVerifiedAt = nil
	}

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	s.logger.Info("User profile updated", logger.Fields{
		"user_id": userID.String(),
	})

	response := user.ToResponse()
	return &response, nil
}

func (s *authService) ChangePassword(ctx context.Context, userID uuid.UUID, req auth.ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	err = s.bcryptService.ComparePassword(user.Password, req.CurrentPassword)
	if err != nil {
		return auth.ErrInvalidCredentials
	}

	err = s.bcryptService.ComparePassword(user.Password, req.NewPassword)
	if err == nil {
		return auth.ErrSamePassword
	}

	hashedPassword, err := s.bcryptService.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	err = s.userRepo.UpdatePassword(ctx, userID, hashedPassword)
	if err != nil {
		return err
	}

	s.logger.Info("User password changed", logger.Fields{
		"user_id": userID.String(),
	})

	return nil
}

func (s *authService) ForgotPassword(ctx context.Context, req auth.ForgotPasswordRequest) error {

	_, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if err == auth.ErrUserNotFound {

			s.logger.Info("Password reset requested for non-existent email", logger.Fields{
				"email": req.Email,
			})
			return nil
		}
		return err
	}

	return s.SendOTP(ctx, auth.SendOTPRequest{
		Email: req.Email,
		Type:  "password_reset",
	})
}

func (s *authService) ResetPassword(ctx context.Context, req auth.ResetPasswordRequest) error {

	err := s.VerifyOTP(ctx, auth.VerifyOTPRequest{
		Email: req.Email,
		OTP:   req.OTP,
		Type:  "password_reset",
	})
	if err != nil {
		return err
	}

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return err
	}

	hashedPassword, err := s.bcryptService.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	err = s.userRepo.UpdatePassword(ctx, user.ID, hashedPassword)
	if err != nil {
		return err
	}

	s.logger.Info("User password reset", logger.Fields{
		"user_id": user.ID.String(),
		"email":   user.Email,
	})

	return nil
}

func (s *authService) SendOTP(ctx context.Context, req auth.SendOTPRequest) error {

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && req.Type != "password_reset" {

		return err
	}

	userName := "User"
	if user != nil {
		userName = user.Name
	}

	otp := s.utilsService.GenerateNumericOTP(6)

	otpKey := fmt.Sprintf("otp:%s:%s", req.Type, req.Email)
	err = s.redisService.SetOTP(ctx, otpKey, otp, 10*time.Minute)
	if err != nil {
		s.logger.Error("Failed to store OTP", logger.Fields{"error": err.Error()})
		return err
	}

	err = s.emailService.SendOTPEmail(req.Email, userName, otp)
	if err != nil {
		s.logger.Error("Failed to send OTP email", logger.Fields{
			"error": err.Error(),
			"email": req.Email,
			"type":  req.Type,
		})
		return fmt.Errorf("failed to send OTP email: %w", err)
	}

	s.logger.Info("OTP sent successfully via email", logger.Fields{
		"email": req.Email,
		"type":  req.Type,
	})

	return nil
}

func (s *authService) VerifyOTP(ctx context.Context, req auth.VerifyOTPRequest) error {

	otpKey := fmt.Sprintf("otp:%s:%s", req.Type, req.Email)
	storedOTP, err := s.redisService.GetOTP(ctx, otpKey)
	if err != nil {
		return auth.ErrOTPNotFound
	}

	if storedOTP != req.OTP {
		return auth.ErrInvalidOTP
	}

	err = s.redisService.DeleteOTP(ctx, otpKey)
	if err != nil {
		s.logger.Error("Failed to delete OTP", logger.Fields{"error": err.Error()})

	}

	if req.Type == "email_verification" {
		user, err := s.userRepo.GetByEmail(ctx, req.Email)
		if err != nil {
			return err
		}

		err = s.userRepo.MarkEmailVerified(ctx, user.ID)
		if err != nil {
			s.logger.Error("Failed to mark email verified", logger.Fields{"error": err.Error()})
			return err
		}

		s.logger.Info("Email verified successfully", logger.Fields{
			"user_id": user.ID.String(),
			"email":   user.Email,
		})
	}

	return nil
}

func (s *authService) UploadProfileImage(ctx context.Context, userID uuid.UUID, file *multipart.FileHeader) (*auth.UploadProfileImageResponse, error) {

	if !s.utilsService.IsImageFile(file.Filename) {
		return nil, auth.ErrInvalidImageFormat
	}

	if file.Size > 5*1024*1024 {
		return nil, auth.ErrImageTooLarge
	}

	filename := s.utilsService.GenerateUniqueFilename(file.Filename)
	key := fmt.Sprintf("profile-images/%s", filename)

	result, err := s.s3Service.UploadFile(file, key)
	if err != nil {
		s.logger.Error("Failed to upload profile image", logger.Fields{"error": err.Error()})
		return nil, err
	}

	err = s.userRepo.UpdateProfileImage(ctx, userID, result.URL)
	if err != nil {
		s.logger.Error("Failed to update profile image URL", logger.Fields{"error": err.Error()})

		s.s3Service.DeleteFile(key)
		return nil, err
	}

	s.logger.Info("Profile image uploaded successfully", logger.Fields{
		"user_id":   userID.String(),
		"image_url": result.URL,
	})

	return &auth.UploadProfileImageResponse{
		ImageURL: result.URL,
		Message:  "Profile image uploaded successfully",
	}, nil
}
