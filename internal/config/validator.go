package config

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func NewValidator() *validator.Validate {
	validate := validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	registerCustomValidations(validate)

	return validate
}

func registerCustomValidations(validate *validator.Validate) {

	validate.RegisterValidation("phone_id", validateIndonesianPhone)

	validate.RegisterValidation("email_required", validateEmailRequired)

	validate.RegisterValidation("meal_type", validateMealType)

	validate.RegisterValidation("day_of_week", validateDayOfWeek)

	validate.RegisterValidation("plan_type", validatePlanType)

	validate.RegisterValidation("strong_password", validateStrongPassword)
}

func validateEmailRequired(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	if email == "" {
		return false
	}
	return validateEmail(email)
}

func validateEmail(email string) bool {

	if len(email) < 5 || len(email) > 254 {
		return false
	}

	atIndex := strings.LastIndex(email, "@")
	if atIndex < 1 || atIndex == len(email)-1 {
		return false
	}

	localPart := email[:atIndex]
	domainPart := email[atIndex+1:]

	if len(localPart) > 64 {
		return false
	}

	if len(domainPart) > 255 {
		return false
	}

	if !strings.Contains(domainPart, ".") {
		return false
	}

	validEmailChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789._%+-"
	for _, char := range email {
		if !strings.ContainsRune(validEmailChars+"@", char) {
			return false
		}
	}

	return true
}

func validateIndonesianPhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()

	if phone == "" {
		return true
	}

	if len(phone) < 10 || len(phone) > 15 {
		return false
	}

	if strings.HasPrefix(phone, "+62") {
		return len(phone) >= 13 && len(phone) <= 16
	}
	if strings.HasPrefix(phone, "62") {
		return len(phone) >= 12 && len(phone) <= 15
	}
	if strings.HasPrefix(phone, "0") {
		return len(phone) >= 10 && len(phone) <= 13
	}

	return false
}

func validateMealType(fl validator.FieldLevel) bool {
	mealType := strings.ToLower(fl.Field().String())
	validTypes := []string{"breakfast", "lunch", "dinner"}

	for _, valid := range validTypes {
		if mealType == valid {
			return true
		}
	}
	return false
}

func validateDayOfWeek(fl validator.FieldLevel) bool {
	day := strings.ToLower(fl.Field().String())
	validDays := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}

	for _, valid := range validDays {
		if day == valid {
			return true
		}
	}
	return false
}

func validatePlanType(fl validator.FieldLevel) bool {
	planType := strings.ToLower(fl.Field().String())
	validPlans := []string{"diet", "protein", "royal"}

	for _, valid := range validPlans {
		if planType == valid {
			return true
		}
	}
	return false
}

func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}
