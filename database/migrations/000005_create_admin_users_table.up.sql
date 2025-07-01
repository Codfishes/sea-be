CREATE TABLE IF NOT EXISTS admin_users (
                                           id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'admin' CHECK (role IN ('admin', 'super_admin', 'moderator')),
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
    );

CREATE INDEX idx_admin_users_email ON admin_users(email);
CREATE INDEX idx_admin_users_role ON admin_users(role);

COMMENT ON TABLE admin_users IS 'Administrative users for SEA Catering management';
COMMENT ON COLUMN admin_users.role IS 'Admin role: admin, super_admin, or moderator';

INSERT INTO public.admin_users(
    id, email, name, password, role, created_at, updated_at)
VALUES ('1', 'nandanatyon@gmail.com', 'tyo', '$2a$12$AmDzX7ttfmfd8uHm8L.8sOcAjQfyeDOLVjWB6/xQZACd1tAYvYxme', 'admin', '2025-06-27 20:36:19.286601', '2025-06-27 20:36:19.286601');
