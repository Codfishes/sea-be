package entity

import "time"

type Testimonial struct {
	ID           string    `db:"id" json:"id"`
	CustomerName string    `db:"customer_name" json:"customer_name"`
	Message      string    `db:"message" json:"message"`
	Rating       int       `db:"rating" json:"rating"`
	IsApproved   bool      `db:"is_approved" json:"is_approved"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}
