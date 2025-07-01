CREATE TABLE IF NOT EXISTS subscription_audit (
                                                  id VARCHAR(36) PRIMARY KEY,
    subscription_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    old_status VARCHAR(50),
    new_status VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    reason VARCHAR(500),
    admin_id VARCHAR(36),
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT fk_subscription_audit_subscription FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE CASCADE,
    CONSTRAINT fk_subscription_audit_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_subscription_audit_admin FOREIGN KEY (admin_id) REFERENCES admin_users(id) ON DELETE SET NULL,
    CONSTRAINT chk_subscription_audit_status CHECK (
                                                       new_status IN ('active', 'paused', 'cancelled')
    ),
    CONSTRAINT chk_subscription_audit_action CHECK (
                                                       action IN ('created', 'paused', 'resumed', 'cancelled', 'reactivated', 'updated')
    )
    );