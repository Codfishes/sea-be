-- Seed: Admin Users
-- Description: Create default admin users for SEA Catering

-- Default super admin (password: admin123!)
INSERT INTO admin_users (id, email, name, password, role, created_at, updated_at) VALUES
                                                                                      ('01HSEACATERING001', 'admin@seacatering.com', 'SEA Catering Admin', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'super_admin', NOW(), NOW()),
                                                                                      ('01HSEACATERING002', 'manager@seacatering.com', 'Brian Manager', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', NOW(), NOW());
