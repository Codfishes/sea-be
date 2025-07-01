// internal/api/subscriptions/dto.go
package subscriptions

import (
	"sea-catering-backend/internal/entity"
	"time"
)

type CreateSubscriptionRequest struct {
	Name         string               `json:"name" validate:"required,min=2,max=100"`
	PhoneNumber  string               `json:"phone_number,omitempty" validate:"omitempty,phone_id"`
	MealPlanID   string               `json:"meal_plan_id" validate:"required"`
	MealTypes    []entity.MealType    `json:"meal_types" validate:"required,min=1"`
	DeliveryDays []entity.DeliveryDay `json:"delivery_days" validate:"required,min=1"`
	Allergies    string               `json:"allergies,omitempty"`
}

type SubscriptionResponse struct {
	entity.SubscriptionWithDetails
	Message string `json:"message,omitempty"`
	Status  string `json:"status"`
}

type PauseSubscriptionRequest struct {
	StartDate time.Time `json:"start_date" validate:"required"`
	EndDate   time.Time `json:"end_date" validate:"required"`
}

type SubscriptionStatsResponse struct {
	TotalSubscriptions     int            `json:"total_subscriptions"`
	ActiveSubscriptions    int            `json:"active_subscriptions"`
	PausedSubscriptions    int            `json:"paused_subscriptions"`
	CancelledSubscriptions int            `json:"cancelled_subscriptions"`
	MonthlyRevenue         float64        `json:"monthly_revenue"`
	NewSubscriptions       int            `json:"new_subscriptions"`
	Reactivations          int            `json:"reactivations"`
	SubscriptionsByPlan    map[string]int `json:"subscriptions_by_plan"`
}

// Helper structs for responses
type SubscriptionListResponse struct {
	Subscriptions []entity.SubscriptionWithDetails `json:"subscriptions"`
	Meta          *PaginationMeta                  `json:"meta,omitempty"`
}

type PaginationMeta struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// Success response structures
type CreateSubscriptionResponse struct {
	Success      bool                           `json:"success"`
	Message      string                         `json:"message"`
	Subscription entity.SubscriptionWithDetails `json:"subscription"`
	NextSteps    []string                       `json:"next_steps"`
}

type ReactivateSubscriptionResponse struct {
	Success      bool                           `json:"success"`
	Message      string                         `json:"message"`
	Subscription entity.SubscriptionWithDetails `json:"subscription"`
	NextSteps    []string                       `json:"next_steps"`
}

type PauseSubscriptionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Details struct {
		PauseStart string `json:"pause_start"`
		PauseEnd   string `json:"pause_end"`
	} `json:"details"`
}

type CancelSubscriptionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Note    string `json:"note"`
}

type UpdateSubscriptionResponse struct {
	Success      bool                           `json:"success"`
	Message      string                         `json:"message"`
	Subscription entity.SubscriptionWithDetails `json:"subscription"`
}
