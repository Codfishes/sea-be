// internal/api/subscriptions/error.go
package subscriptions

import "errors"

var (
	ErrSubscriptionNotFound      = errors.New("subscription not found")
	ErrInvalidMealPlan           = errors.New("invalid meal plan")
	ErrInvalidMealTypes          = errors.New("invalid meal types")
	ErrInvalidDeliveryDays       = errors.New("invalid delivery days")
	ErrSubscriptionAlreadyExists = errors.New("subscription already exists")
	ErrSubscriptionCancelled     = errors.New("subscription is cancelled")
	ErrSubscriptionPaused        = errors.New("subscription is paused")
	ErrInvalidPauseDates         = errors.New("invalid pause dates")
	ErrUnauthorizedAccess        = errors.New("unauthorized access to subscription")
	ErrSubscriptionActive        = errors.New("subscription is already active")
	ErrInvalidSubscriptionStatus = errors.New("invalid subscription status for this operation")
	ErrInvalidDateRange          = errors.New("invalid date range provided")
	ErrSubscriptionUpdateFailed  = errors.New("failed to update subscription")
)

// HTTP Status Code mappings
func GetHTTPStatusCode(err error) int {
	switch err {
	case ErrSubscriptionNotFound:
		return 404
	case ErrInvalidMealPlan, ErrInvalidMealTypes, ErrInvalidDeliveryDays,
		ErrInvalidPauseDates, ErrInvalidSubscriptionStatus, ErrInvalidDateRange:
		return 400
	case ErrUnauthorizedAccess:
		return 403
	case ErrSubscriptionAlreadyExists:
		return 409
	case ErrSubscriptionUpdateFailed:
		return 422
	default:
		return 500
	}
}

// Error message mappings
func GetErrorMessage(err error) string {
	switch err {
	case ErrSubscriptionNotFound:
		return "Subscription not found"
	case ErrInvalidMealPlan:
		return "Selected meal plan is invalid or inactive"
	case ErrInvalidMealTypes:
		return "Invalid meal types selected. Please choose from: breakfast, lunch, dinner"
	case ErrInvalidDeliveryDays:
		return "Invalid delivery days selected. Please choose valid weekdays"
	case ErrSubscriptionAlreadyExists:
		return "You already have an active subscription for this meal plan"
	case ErrSubscriptionCancelled:
		return "This subscription has been cancelled"
	case ErrSubscriptionPaused:
		return "This subscription is currently paused"
	case ErrSubscriptionActive:
		return "This subscription is already active"
	case ErrInvalidPauseDates:
		return "Invalid pause dates. Start date must be in the future and before end date"
	case ErrUnauthorizedAccess:
		return "You don't have permission to access this subscription"
	case ErrInvalidSubscriptionStatus:
		return "Current subscription status doesn't allow this operation"
	case ErrInvalidDateRange:
		return "Invalid date range provided"
	case ErrSubscriptionUpdateFailed:
		return "Failed to update subscription. Please try again"
	default:
		return "An unexpected error occurred"
	}
}
