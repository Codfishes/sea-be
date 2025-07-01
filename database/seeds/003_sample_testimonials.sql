-- Seed: Sample Testimonials
-- Description: Create sample approved testimonials

INSERT INTO testimonials (id, customer_name, message, rating, is_approved, created_at, updated_at) VALUES
                                                                                                       (
                                                                                                           '01HSEACATERING201',
                                                                                                           'Sarah Johnson',
                                                                                                           'SEA Catering has completely transformed my eating habits! The Diet Plan is perfectly portioned and incredibly delicious. I''ve lost 5kg in just 2 months!',
                                                                                                           5,
                                                                                                           true,
                                                                                                           NOW() - INTERVAL '30 days',
                                                                                                           NOW() - INTERVAL '30 days'
                                                                                                       ),
                                                                                                       (
                                                                                                           '01HSEACATERING202',
                                                                                                           'Muhammad Rizki',
                                                                                                           'As a busy professional, SEA Catering is a lifesaver. The delivery is always on time and the food quality is exceptional. Highly recommended!',
                                                                                                           5,
                                                                                                           true,
                                                                                                           NOW() - INTERVAL '25 days',
                                                                                                           NOW() - INTERVAL '25 days'
                                                                                                       ),
                                                                                                       (
                                                                                                           '01HSEACATERING203',
                                                                                                           'Amanda Putri',
                                                                                                           'The Protein Plan has been perfect for my fitness journey. The meals are tasty and help me reach my daily protein goals effortlessly.',
                                                                                                           4,
                                                                                                           true,
                                                                                                           NOW() - INTERVAL '20 days',
                                                                                                           NOW() - INTERVAL '20 days'
                                                                                                       ),
                                                                                                       (
                                                                                                           '01HSEACATERING204',
                                                                                                           'David Chen',
                                                                                                           'Royal Plan is absolutely amazing! The gourmet meals feel like dining at a 5-star restaurant. Worth every penny!',
                                                                                                           5,
                                                                                                           true,
                                                                                                           NOW() - INTERVAL '15 days',
                                                                                                           NOW() - INTERVAL '15 days'
                                                                                                       ),
                                                                                                       (
                                                                                                           '01HSEACATERING205',
                                                                                                           'Siti Nurhaliza',
                                                                                                           'Great service and healthy options. The customization for my allergies was handled perfectly. Thank you SEA Catering!',
                                                                                                           4,
                                                                                                           true,
                                                                                                           NOW() - INTERVAL '10 days',
                                                                                                           NOW() - INTERVAL '10 days'
                                                                                                       );