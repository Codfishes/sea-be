package meal_plans

import (
	"time"
)

type MealPlanResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	ImageURL    string    `json:"image_url,omitempty"`
	Features    []string  `json:"features"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateMealPlanRequest struct {
	Name        string   `json:"name" validate:"required,min=2,max=100"`
	Description string   `json:"description" validate:"required,min=10,max=500"`
	Price       float64  `json:"price" validate:"required,min=0"`
	ImageURL    string   `json:"image_url,omitempty" validate:"omitempty,url"`
	Features    []string `json:"features" validate:"required,min=1"`
}

type UpdateMealPlanRequest struct {
	Name        *string   `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Description *string   `json:"description,omitempty" validate:"omitempty,min=10,max=500"`
	Price       *float64  `json:"price,omitempty" validate:"omitempty,min=0"`
	ImageURL    *string   `json:"image_url,omitempty" validate:"omitempty,url"`
	Features    *[]string `json:"features,omitempty" validate:"omitempty,min=1"`
	IsActive    *bool     `json:"is_active,omitempty"`
}

type MealPlanListRequest struct {
	Page     int    `query:"page" validate:"omitempty,min=1"`
	Limit    int    `query:"limit" validate:"omitempty,min=1,max=100"`
	Search   string `query:"search" validate:"omitempty,max=100"`
	IsActive *bool  `query:"is_active"`
	SortBy   string `query:"sort_by" validate:"omitempty,oneof=name price created_at updated_at"`
	SortDir  string `query:"sort_dir" validate:"omitempty,oneof=asc desc"`
}

type MealPlanListResponse struct {
	MealPlans []MealPlanResponse `json:"meal_plans"`
	Meta      *PaginationMeta    `json:"meta"`
}

type PaginationMeta struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

type MealPlanStatsResponse struct {
	TotalMealPlans    int     `json:"total_meal_plans"`
	ActiveMealPlans   int     `json:"active_meal_plans"`
	InactiveMealPlans int     `json:"inactive_meal_plans"`
	AveragePrice      float64 `json:"average_price"`
	MostPopularPlan   string  `json:"most_popular_plan"`
	TotalSubscribers  int     `json:"total_subscribers"`
}
