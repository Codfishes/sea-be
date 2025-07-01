package context

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ContextKey string

const (
	RequestIDKey     ContextKey = "request_id"
	UserIDKey        ContextKey = "user_id"
	UserEmailKey     ContextKey = "user_email"
	UserRoleKey      ContextKey = "user_role"
	TraceIDKey       ContextKey = "trace_id"
	SpanIDKey        ContextKey = "span_id"
	IPAddressKey     ContextKey = "ip_address"
	UserAgentKey     ContextKey = "user_agent"
	SessionIDKey     ContextKey = "session_id"
	TenantIDKey      ContextKey = "tenant_id"
	CorrelationIDKey ContextKey = "correlation_id"
)

func WithRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		requestID = generateID()
	}
	return context.WithValue(ctx, RequestIDKey, requestID)
}

func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

func WithUserEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, UserEmailKey, email)
}

func GetUserEmail(ctx context.Context) string {
	if email, ok := ctx.Value(UserEmailKey).(string); ok {
		return email
	}
	return ""
}

func WithUserRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, UserRoleKey, role)
}

func GetUserRole(ctx context.Context) string {
	if role, ok := ctx.Value(UserRoleKey).(string); ok {
		return role
	}
	return ""
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	if traceID == "" {
		traceID = generateID()
	}
	return context.WithValue(ctx, TraceIDKey, traceID)
}

func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

func WithSpanID(ctx context.Context, spanID string) context.Context {
	if spanID == "" {
		spanID = generateID()
	}
	return context.WithValue(ctx, SpanIDKey, spanID)
}

func GetSpanID(ctx context.Context) string {
	if spanID, ok := ctx.Value(SpanIDKey).(string); ok {
		return spanID
	}
	return ""
}

func WithIPAddress(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, IPAddressKey, ip)
}

func GetIPAddress(ctx context.Context) string {
	if ip, ok := ctx.Value(IPAddressKey).(string); ok {
		return ip
	}
	return ""
}

func WithUserAgent(ctx context.Context, userAgent string) context.Context {
	return context.WithValue(ctx, UserAgentKey, userAgent)
}

func GetUserAgent(ctx context.Context) string {
	if userAgent, ok := ctx.Value(UserAgentKey).(string); ok {
		return userAgent
	}
	return ""
}

func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, SessionIDKey, sessionID)
}

func GetSessionID(ctx context.Context) string {
	if sessionID, ok := ctx.Value(SessionIDKey).(string); ok {
		return sessionID
	}
	return ""
}

func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

func GetTenantID(ctx context.Context) string {
	if tenantID, ok := ctx.Value(TenantIDKey).(string); ok {
		return tenantID
	}
	return ""
}

func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	if correlationID == "" {
		correlationID = generateID()
	}
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

func GetCorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return correlationID
	}
	return ""
}

func FromFiberContext(c *fiber.Ctx) context.Context {
	ctx := context.Background()

	requestID := c.Get("X-Request-ID")
	if requestID == "" {
		requestID = generateID()
	}
	ctx = WithRequestID(ctx, requestID)

	if traceID := c.Get("X-Trace-ID"); traceID != "" {
		ctx = WithTraceID(ctx, traceID)
	}

	if correlationID := c.Get("X-Correlation-ID"); correlationID != "" {
		ctx = WithCorrelationID(ctx, correlationID)
	}

	ctx = WithIPAddress(ctx, c.IP())

	ctx = WithUserAgent(ctx, c.Get("User-Agent"))

	if userID := c.Locals("user_id"); userID != nil {
		if id, ok := userID.(string); ok {
			ctx = WithUserID(ctx, id)
		}
	}

	if userEmail := c.Locals("user_email"); userEmail != nil {
		if email, ok := userEmail.(string); ok {
			ctx = WithUserEmail(ctx, email)
		}
	}

	if userRole := c.Locals("user_role"); userRole != nil {
		if role, ok := userRole.(string); ok {
			ctx = WithUserRole(ctx, role)
		}
	}

	if sessionID := c.Locals("session_id"); sessionID != nil {
		if id, ok := sessionID.(string); ok {
			ctx = WithSessionID(ctx, id)
		}
	}

	return ctx
}

func ToFiberContext(c *fiber.Ctx, ctx context.Context) {
	if requestID := GetRequestID(ctx); requestID != "" {
		c.Locals("request_id", requestID)
		c.Set("X-Request-ID", requestID)
	}

	if traceID := GetTraceID(ctx); traceID != "" {
		c.Locals("trace_id", traceID)
		c.Set("X-Trace-ID", traceID)
	}

	if correlationID := GetCorrelationID(ctx); correlationID != "" {
		c.Locals("correlation_id", correlationID)
		c.Set("X-Correlation-ID", correlationID)
	}

	if userID := GetUserID(ctx); userID != "" {
		c.Locals("user_id", userID)
	}

	if userEmail := GetUserEmail(ctx); userEmail != "" {
		c.Locals("user_email", userEmail)
	}

	if userRole := GetUserRole(ctx); userRole != "" {
		c.Locals("user_role", userRole)
	}

	if sessionID := GetSessionID(ctx); sessionID != "" {
		c.Locals("session_id", sessionID)
	}

	if tenantID := GetTenantID(ctx); tenantID != "" {
		c.Locals("tenant_id", tenantID)
	}
}

func WithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, timeout)
}

func WithDeadline(parent context.Context, deadline time.Time) (context.Context, context.CancelFunc) {
	return context.WithDeadline(parent, deadline)
}

func WithCancel(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithCancel(parent)
}

func Background() context.Context {
	ctx := context.Background()
	return WithRequestID(ctx, generateID())
}

func TODO() context.Context {
	ctx := context.TODO()
	return WithRequestID(ctx, generateID())
}

func IsContextCanceled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func GetContextError(ctx context.Context) error {
	return ctx.Err()
}

func WaitForContext(ctx context.Context, timeout time.Duration) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	<-timeoutCtx.Done()
	return timeoutCtx.Err()
}

func CreateChildContext(parent context.Context) context.Context {
	child := context.Background()

	if requestID := GetRequestID(parent); requestID != "" {
		child = WithRequestID(child, requestID)
	}

	if traceID := GetTraceID(parent); traceID != "" {
		child = WithTraceID(child, traceID)
	}

	if correlationID := GetCorrelationID(parent); correlationID != "" {
		child = WithCorrelationID(child, correlationID)
	}

	if userID := GetUserID(parent); userID != "" {
		child = WithUserID(child, userID)
	}

	if userEmail := GetUserEmail(parent); userEmail != "" {
		child = WithUserEmail(child, userEmail)
	}

	if userRole := GetUserRole(parent); userRole != "" {
		child = WithUserRole(child, userRole)
	}

	if sessionID := GetSessionID(parent); sessionID != "" {
		child = WithSessionID(child, sessionID)
	}

	if tenantID := GetTenantID(parent); tenantID != "" {
		child = WithTenantID(child, tenantID)
	}

	if ipAddress := GetIPAddress(parent); ipAddress != "" {
		child = WithIPAddress(child, ipAddress)
	}

	if userAgent := GetUserAgent(parent); userAgent != "" {
		child = WithUserAgent(child, userAgent)
	}

	return child
}

func ExtractMetadata(ctx context.Context) map[string]interface{} {
	metadata := make(map[string]interface{})

	if requestID := GetRequestID(ctx); requestID != "" {
		metadata["request_id"] = requestID
	}

	if traceID := GetTraceID(ctx); traceID != "" {
		metadata["trace_id"] = traceID
	}

	if correlationID := GetCorrelationID(ctx); correlationID != "" {
		metadata["correlation_id"] = correlationID
	}

	if userID := GetUserID(ctx); userID != "" {
		metadata["user_id"] = userID
	}

	if userEmail := GetUserEmail(ctx); userEmail != "" {
		metadata["user_email"] = userEmail
	}

	if userRole := GetUserRole(ctx); userRole != "" {
		metadata["user_role"] = userRole
	}

	if sessionID := GetSessionID(ctx); sessionID != "" {
		metadata["session_id"] = sessionID
	}

	if tenantID := GetTenantID(ctx); tenantID != "" {
		metadata["tenant_id"] = tenantID
	}

	if ipAddress := GetIPAddress(ctx); ipAddress != "" {
		metadata["ip_address"] = ipAddress
	}

	if userAgent := GetUserAgent(ctx); userAgent != "" {
		metadata["user_agent"] = userAgent
	}

	if spanID := GetSpanID(ctx); spanID != "" {
		metadata["span_id"] = spanID
	}

	return metadata
}

func CreateContextWithMetadata(metadata map[string]interface{}) context.Context {
	ctx := context.Background()

	if requestID, ok := metadata["request_id"].(string); ok && requestID != "" {
		ctx = WithRequestID(ctx, requestID)
	}

	if traceID, ok := metadata["trace_id"].(string); ok && traceID != "" {
		ctx = WithTraceID(ctx, traceID)
	}

	if correlationID, ok := metadata["correlation_id"].(string); ok && correlationID != "" {
		ctx = WithCorrelationID(ctx, correlationID)
	}

	if userID, ok := metadata["user_id"].(string); ok && userID != "" {
		ctx = WithUserID(ctx, userID)
	}

	if userEmail, ok := metadata["user_email"].(string); ok && userEmail != "" {
		ctx = WithUserEmail(ctx, userEmail)
	}

	if userRole, ok := metadata["user_role"].(string); ok && userRole != "" {
		ctx = WithUserRole(ctx, userRole)
	}

	if sessionID, ok := metadata["session_id"].(string); ok && sessionID != "" {
		ctx = WithSessionID(ctx, sessionID)
	}

	if tenantID, ok := metadata["tenant_id"].(string); ok && tenantID != "" {
		ctx = WithTenantID(ctx, tenantID)
	}

	if ipAddress, ok := metadata["ip_address"].(string); ok && ipAddress != "" {
		ctx = WithIPAddress(ctx, ipAddress)
	}

	if userAgent, ok := metadata["user_agent"].(string); ok && userAgent != "" {
		ctx = WithUserAgent(ctx, userAgent)
	}

	if spanID, ok := metadata["span_id"].(string); ok && spanID != "" {
		ctx = WithSpanID(ctx, spanID)
	}

	return ctx
}

func IsUserAuthenticated(ctx context.Context) bool {
	return GetUserID(ctx) != ""
}

func IsUserAdmin(ctx context.Context) bool {
	return GetUserRole(ctx) == "admin"
}

func HasRole(ctx context.Context, role string) bool {
	return GetUserRole(ctx) == role
}

func GetContextDuration(ctx context.Context, startTime time.Time) time.Duration {
	return time.Since(startTime)
}

func generateID() string {
	return uuid.New().String()
}
