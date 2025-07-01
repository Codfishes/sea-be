package jwt

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

type Interface interface {
	GenerateAccessToken(userID string, email string, role string) (AccessTokenResponse, error)
	ValidateAccessToken(tokenString string) (*Claims, error)
	ExtractTokenFromHeader(authHeader string) (string, error)
	ExtractTokenFromFiberContext(c *fiber.Ctx) (string, error)
	RevokeToken(ctx context.Context, tokenString string) error
	IsTokenRevoked(ctx context.Context, tokenString string) (bool, error)
	CleanupExpiredTokens(ctx context.Context) error
}

type Service struct {
	config      *Config
	redisClient *redis.Client
}

type Config struct {
	AccessTokenSecret string
	AccessTokenExpiry time.Duration
	Issuer            string
	Algorithm         string
}

type Claims struct {
	UserID     string `json:"user_id"`
	Email      string `json:"email"`
	IsVerified bool   `json:"is_verified,omitempty"`
	Role       string `json:"role"`
	jwt.RegisteredClaims
}

type AccessTokenResponse struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func LoadConfig() *Config {
	accessSecret := os.Getenv("JWT_SECRET")
	if accessSecret == "" {
		accessSecret = "default-secret-change-in-production"
	}

	accessExpiry := 24 * time.Hour
	if envExpiry := os.Getenv("JWT_EXPIRES_IN"); envExpiry != "" {
		if parsed, err := time.ParseDuration(envExpiry); err == nil {
			accessExpiry = parsed
		}
	}

	issuer := os.Getenv("JWT_ISSUER")
	if issuer == "" {
		issuer = "sea-catering-backend"
	}

	return &Config{
		AccessTokenSecret: accessSecret,
		AccessTokenExpiry: accessExpiry,
		Issuer:            issuer,
		Algorithm:         "HS256",
	}
}

func New() Interface {
	config := LoadConfig()
	return NewWithConfig(config, nil)
}

func NewWithConfig(config *Config, redisClient *redis.Client) Interface {
	if config == nil {
		config = LoadConfig()
	}

	return &Service{
		config:      config,
		redisClient: redisClient,
	}
}

func (s *Service) GenerateAccessToken(userID string, email string, role string) (AccessTokenResponse, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.AccessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.config.Issuer,
			Subject:   userID,
			ID:        generateJTI(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	accessToken, err := token.SignedString([]byte(s.config.AccessTokenSecret))
	if err != nil {
		return AccessTokenResponse{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return AccessTokenResponse{
		AccessToken: accessToken,
		ExpiresAt:   claims.ExpiresAt.Time,
	}, nil
}

func (s *Service) ValidateAccessToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token is empty")
	}

	if s.redisClient != nil {
		ctx := context.Background()
		revoked, err := s.IsTokenRevoked(ctx, tokenString)
		if err != nil {

		} else if revoked {
			return nil, fmt.Errorf("token has been revoked")
		}
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.AccessTokenSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func (s *Service) ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is empty")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return parts[1], nil
}

func (s *Service) ExtractTokenFromFiberContext(c *fiber.Ctx) (string, error) {
	authHeader := c.Get("Authorization")
	return s.ExtractTokenFromHeader(authHeader)
}

func (s *Service) RevokeToken(ctx context.Context, tokenString string) error {
	if tokenString == "" {
		return fmt.Errorf("token is empty")
	}

	if s.redisClient == nil {
		return fmt.Errorf("Redis client not available for token revocation")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		return []byte("dummy"), nil
	})

	var ttl time.Duration = s.config.AccessTokenExpiry

	if err == nil {
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if exp, ok := claims["exp"].(float64); ok {
				expTime := time.Unix(int64(exp), 0)
				if time.Now().Before(expTime) {
					ttl = time.Until(expTime)
				}
			}
		}
	}

	revokedKey := fmt.Sprintf("revoked_token:%s", generateTokenHash(tokenString))
	err = s.redisClient.Set(ctx, revokedKey, time.Now().Unix(), ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to revoke token in Redis: %w", err)
	}

	return nil
}

func (s *Service) IsTokenRevoked(ctx context.Context, tokenString string) (bool, error) {
	if s.redisClient == nil {
		return false, nil
	}

	revokedKey := fmt.Sprintf("revoked_token:%s", generateTokenHash(tokenString))
	exists, err := s.redisClient.Exists(ctx, revokedKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check token revocation: %w", err)
	}

	return exists > 0, nil
}

func (s *Service) CleanupExpiredTokens(ctx context.Context) error {
	if s.redisClient == nil {
		return nil
	}

	return nil
}

func generateJTI() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func generateTokenHash(token string) string {

	return fmt.Sprintf("%x", len(token)*31+int(token[0]))
}

func GetUserFromToken(c *fiber.Ctx) (*Claims, error) {
	user := c.Locals("user")
	if user == nil {
		return nil, fmt.Errorf("user not found in context")
	}

	claims, ok := user.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid user claims in context")
	}

	return claims, nil
}

func GetUserID(c *fiber.Ctx) (string, error) {
	claims, err := GetUserFromToken(c)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

func GetUserEmail(c *fiber.Ctx) (string, error) {
	claims, err := GetUserFromToken(c)
	if err != nil {
		return "", err
	}
	return claims.Email, nil
}

func GetUserRole(c *fiber.Ctx) (string, error) {
	claims, err := GetUserFromToken(c)
	if err != nil {
		return "", err
	}
	return claims.Role, nil
}

func IsAdmin(c *fiber.Ctx) (bool, error) {
	role, err := GetUserRole(c)
	if err != nil {
		return false, err
	}
	return strings.ToLower(role) == "admin", nil
}
