package entity

import "time"

type MealType string
type DeliveryDay string
type SubscriptionStatus string

const (
	MealTypeBreakfast MealType = "breakfast"
	MealTypeLunch     MealType = "lunch"
	MealTypeDinner    MealType = "dinner"
)

const (
	DayMonday    DeliveryDay = "monday"
	DayTuesday   DeliveryDay = "tuesday"
	DayWednesday DeliveryDay = "wednesday"
	DayThursday  DeliveryDay = "thursday"
	DayFriday    DeliveryDay = "friday"
	DaySaturday  DeliveryDay = "saturday"
	DaySunday    DeliveryDay = "sunday"
)

const (
	StatusActive    SubscriptionStatus = "active"
	StatusPaused    SubscriptionStatus = "paused"
	StatusCancelled SubscriptionStatus = "cancelled"
)

type Subscription struct {
	ID             string             `db:"id" json:"id"`
	UserID         string             `db:"user_id" json:"user_id"`
	MealPlanID     string             `db:"meal_plan_id" json:"meal_plan_id"`
	MealTypes      []MealType         `db:"meal_types" json:"meal_types"`
	DeliveryDays   []DeliveryDay      `db:"delivery_days" json:"delivery_days"`
	Allergies      string             `db:"allergies" json:"allergies,omitempty"`
	TotalPrice     float64            `db:"total_price" json:"total_price"`
	Status         SubscriptionStatus `db:"status" json:"status"`
	PauseStartDate *time.Time         `db:"pause_start_date" json:"pause_start_date,omitempty"`
	PauseEndDate   *time.Time         `db:"pause_end_date" json:"pause_end_date,omitempty"`
	CreatedAt      time.Time          `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `db:"updated_at" json:"updated_at"`
}

type SubscriptionWithDetails struct {
	Subscription
	MealPlan MealPlan `json:"meal_plan"`
}
