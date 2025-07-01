package testimonials

import "errors"

var (
	ErrTestimonialNotFound = errors.New("testimonial not found")
	ErrInvalidRating       = errors.New("rating must be between 1 and 5")
	ErrTestimonialExists   = errors.New("testimonial already exists")
	ErrUnauthorizedAccess  = errors.New("unauthorized access to testimonial")
)
