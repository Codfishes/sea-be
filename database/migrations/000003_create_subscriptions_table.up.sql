CREATE TYPE meal_type AS ENUM ('breakfast', 'lunch', 'dinner');
CREATE TYPE delivery_day AS ENUM ('monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday');
CREATE TYPE subscription_status AS ENUM ('active', 'paused', 'cancelled');

CREATE TABLE IF NOT EXISTS subscriptions (
                                             id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    meal_plan_id VARCHAR(36) NOT NULL,
    meal_types meal_type[] NOT NULL,
    delivery_days delivery_day[] NOT NULL,
    allergies TEXT,
    total_price DECIMAL(10, 2) NOT NULL CHECK (total_price > 0),
    status subscription_status DEFAULT 'active',
    pause_start_date DATE,
    pause_end_date DATE,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT fk_subscriptions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_subscriptions_meal_plan FOREIGN KEY (meal_plan_id) REFERENCES meal_plans(id),
    CONSTRAINT chk_pause_dates CHECK (
(pause_start_date IS NULL AND pause_end_date IS NULL) OR
(pause_start_date IS NOT NULL AND pause_end_date IS NOT NULL AND pause_end_date > pause_start_date)
    ),
    CONSTRAINT chk_meal_types_not_empty CHECK (array_length(meal_types, 1) > 0),
    CONSTRAINT chk_delivery_days_not_empty CHECK (array_length(delivery_days, 1) > 0)
    );

CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_meal_plan_id ON subscriptions(meal_plan_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_created_at ON subscriptions(created_at);

COMMENT ON TABLE subscriptions IS 'User meal subscriptions';
COMMENT ON COLUMN subscriptions.meal_types IS 'Selected meal types for subscription';
COMMENT ON COLUMN subscriptions.delivery_days IS 'Selected delivery days';
COMMENT ON COLUMN subscriptions.total_price IS 'Calculated total price for subscription';