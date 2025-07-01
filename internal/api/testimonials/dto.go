package testimonials

type CreateTestimonialRequest struct {
	CustomerName string `json:"customer_name" validate:"required,min=2,max=100"`
	Message      string `json:"message" validate:"required,min=10,max=1000"`
	Rating       int    `json:"rating" validate:"required,min=1,max=5"`
}

type TestimonialResponse struct {
	ID           string `json:"id"`
	CustomerName string `json:"customer_name"`
	Message      string `json:"message"`
	Rating       int    `json:"rating"`
	IsApproved   bool   `json:"is_approved"`
	CreatedAt    string `json:"created_at"`
}
