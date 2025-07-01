package bcrypt

import (
	"fmt"
	"os"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

type Interface interface {
	HashPassword(password string) (string, error)
	ComparePassword(hashedPassword, password string) error
	GetCost() int
}

type Service struct {
	cost int
}

type Config struct {
	Cost int
}

func LoadConfig() *Config {
	cost := bcrypt.DefaultCost

	if envCost := os.Getenv("BCRYPT_COST"); envCost != "" {
		if parsed, err := strconv.Atoi(envCost); err == nil {
			if parsed >= bcrypt.MinCost && parsed <= bcrypt.MaxCost {
				cost = parsed
			}
		}
	}

	return &Config{
		Cost: cost,
	}
}

func New() Interface {
	config := LoadConfig()
	return NewWithConfig(config)
}

func NewWithConfig(config *Config) Interface {
	if config == nil {
		config = LoadConfig()
	}

	if config.Cost < bcrypt.MinCost || config.Cost > bcrypt.MaxCost {
		config.Cost = bcrypt.DefaultCost
	}

	return &Service{
		cost: config.Cost,
	}
}

func NewWithCost(cost int) Interface {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}

	return &Service{
		cost: cost,
	}
}

func (s *Service) HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	if len(password) > 72 {
		return "", fmt.Errorf("password too long (max 72 characters)")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

func (s *Service) ComparePassword(hashedPassword, password string) error {
	if hashedPassword == "" {
		return fmt.Errorf("hashed password cannot be empty")
	}

	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return fmt.Errorf("invalid password")
		}
		return fmt.Errorf("failed to compare password: %w", err)
	}

	return nil
}

func (s *Service) GetCost() int {
	return s.cost
}

func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	if len(password) > 72 {
		return fmt.Errorf("password must be at most 72 characters long")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char >= '!' && char <= '/' || char >= ':' && char <= '@' || char >= '[' && char <= '`' || char >= '{' && char <= '~':
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}

	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

func IsHashed(password string) bool {

	if len(password) != 60 {
		return false
	}

	return password[0] == '$' && password[3] == '$' && password[6] == '$' &&
		(password[1:3] == "2a" || password[1:3] == "2b" || password[1:3] == "2x" || password[1:3] == "2y")
}
