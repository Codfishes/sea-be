package meal_plans

import "errors"

var (
	ErrMealPlanNotFound         = errors.New("meal plan not found")
	ErrMealPlanAlreadyExists    = errors.New("meal plan already exists")
	ErrMealPlanInactive         = errors.New("meal plan is inactive")
	ErrInvalidPrice             = errors.New("invalid price value")
	ErrInvalidFeatures          = errors.New("invalid features provided")
	ErrMealPlanHasSubscriptions = errors.New("meal plan has active subscriptions and cannot be deleted")
	ErrInvalidSortField         = errors.New("invalid sort field")
	ErrInvalidImageURL          = errors.New("invalid image URL")
	ErrMealPlanNameTaken        = errors.New("meal plan name is already taken")
	ErrInvalidPaginationParams  = errors.New("invalid pagination parameters")
)

func GetHTTPStatusCode(err error) int {
	switch err {
	case ErrMealPlanNotFound:
		return 404
	case ErrMealPlanAlreadyExists, ErrMealPlanNameTaken:
		return 409
	case ErrMealPlanInactive, ErrInvalidPrice, ErrInvalidFeatures,
		ErrInvalidSortField, ErrInvalidImageURL, ErrInvalidPaginationParams:
		return 400
	case ErrMealPlanHasSubscriptions:
		return 422
	default:
		return 500
	}
}

func GetErrorMessage(err error) string {
	switch err {
	case ErrMealPlanNotFound:
		return "The requested meal plan was not found"
	case ErrMealPlanAlreadyExists:
		return "A meal plan with this name already exists"
	case ErrMealPlanNameTaken:
		return "This meal plan name is already taken"
	case ErrMealPlanInactive:
		return "This meal plan is currently inactive"
	case ErrInvalidPrice:
		return "Price must be a positive number"
	case ErrInvalidFeatures:
		return "At least one feature must be provided"
	case ErrMealPlanHasSubscriptions:
		return "Cannot delete meal plan with active subscriptions"
	case ErrInvalidSortField:
		return "Invalid sort field. Use: name, price, created_at, or updated_at"
	case ErrInvalidImageURL:
		return "Invalid image URL format"
	case ErrInvalidPaginationParams:
		return "Invalid pagination parameters"
	default:
		return "An unexpected error occurred"
	}
}
