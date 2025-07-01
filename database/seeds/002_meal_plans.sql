-- Seed: Meal Plans
-- Description: Create default meal plans for SEA Catering

INSERT INTO meal_plans (id, name, description, price, features, is_active, created_at, updated_at) VALUES
                                                                                                       (
                                                                                                           '01HSEACATERING101',
                                                                                                           'Diet Plan',
                                                                                                           'Healthy meal plan designed for weight management and balanced nutrition. Perfect for those looking to maintain a healthy lifestyle.',
                                                                                                           30000.00,
                                                                                                           ARRAY['Low calorie meals', 'High protein content', 'Nutritionist approved', 'Balanced macronutrients', 'Fresh ingredients'],
                                                                                                           true,
                                                                                                           NOW(),
                                                                                                           NOW()
                                                                                                       ),
                                                                                                       (
                                                                                                           '01HSEACATERING102',
                                                                                                           'Protein Plan',
                                                                                                           'High protein meals designed for fitness enthusiasts and athletes. Ideal for muscle building and post-workout recovery.',
                                                                                                           40000.00,
                                                                                                           ARRAY['High protein content', 'Post-workout meals', 'Muscle building focus', 'Lean meat options', 'Performance nutrition'],
                                                                                                           true,
                                                                                                           NOW(),
                                                                                                           NOW()
                                                                                                       ),
                                                                                                       (
                                                                                                           '01HSEACATERING103',
                                                                                                           'Royal Plan',
                                                                                                           'Premium gourmet meals with the finest ingredients. A luxury dining experience delivered to your door.',
                                                                                                           60000.00,
                                                                                                           ARRAY['Premium ingredients', 'Gourmet recipes', 'Chef curated', 'International cuisine', 'Fine dining experience'],
                                                                                                           true,
                                                                                                           NOW(),
                                                                                                           NOW()
                                                                                                       );