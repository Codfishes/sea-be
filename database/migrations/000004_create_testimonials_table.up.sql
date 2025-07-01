CREATE TABLE IF NOT EXISTS testimonials (
                                            id VARCHAR(36) PRIMARY KEY,
    customer_name VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    is_approved BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
    );

CREATE INDEX idx_testimonials_approved ON testimonials(is_approved);
CREATE INDEX idx_testimonials_rating ON testimonials(rating);
CREATE INDEX idx_testimonials_created_at ON testimonials(created_at);

COMMENT ON TABLE testimonials IS 'Customer testimonials and reviews';
COMMENT ON COLUMN testimonials.rating IS 'Rating from 1 to 5 stars';
COMMENT ON COLUMN testimonials.is_approved IS 'Whether testimonial is approved for display';