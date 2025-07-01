CREATE TABLE IF NOT EXISTS meal_plans (
                                          id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL CHECK (price > 0),
    image_url VARCHAR(500),
    features TEXT[],
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
    );

CREATE INDEX idx_meal_plans_active ON meal_plans(is_active);
CREATE INDEX idx_meal_plans_price ON meal_plans(price);

COMMENT ON TABLE meal_plans IS 'Available meal plans for subscription';
COMMENT ON COLUMN meal_plans.features IS 'Array of meal plan features';
COMMENT ON COLUMN meal_plans.price IS 'Price per meal in IDR';