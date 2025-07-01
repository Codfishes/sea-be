package utils

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math"
	"math/big"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/redis/go-redis/v9"
)

type Interface interface {
	GenerateULID() string
	GenerateULIDWithTime(t time.Time) string
	GenerateRandomString(length int) string
	GenerateNumericOTP(length int) string
	GenerateAlphanumericCode(length int) string

	IsValidEmail(email string) bool
	IsValidPhone(phone string) bool
	IsValidIndonesianPhone(phone string) bool
	IsStrongPassword(password string) bool
	IsValidURL(url string) bool

	Slugify(text string) string
	TruncateString(text string, maxLength int) string
	CapitalizeFirst(text string) string
	CapitalizeWords(text string) string
	RemoveSpaces(text string) string
	NormalizeWhitespace(text string) string

	FormatPrice(amount float64) string
	FormatPriceIDR(amount float64) string
	ParsePrice(priceStr string) (float64, error)
	RoundToDecimal(value float64, places int) float64
	CalculatePercentage(part, total float64) float64

	FormatDateID(t time.Time) string
	FormatDateTimeID(t time.Time) string
	ParseDateID(dateStr string) (time.Time, error)
	GetStartOfDay(t time.Time) time.Time
	GetEndOfDay(t time.Time) time.Time
	GetStartOfWeek(t time.Time) time.Time
	GetEndOfWeek(t time.Time) time.Time
	GetStartOfMonth(t time.Time) time.Time
	GetEndOfMonth(t time.Time) time.Time
	DaysBetween(start, end time.Time) int
	WeeksBetween(start, end time.Time) int
	MonthsBetween(start, end time.Time) int

	CalculateOffset(page, limit int) int
	CalculateTotalPages(totalItems, limit int) int
	ValidatePagination(page, limit int) (int, int)

	CalculateSubscriptionPrice(planPrice float64, mealTypes []string, deliveryDays []string) float64
	ValidateMealTypes(mealTypes []string) error
	ValidateDeliveryDays(deliveryDays []string) error
	GetMealTypeDisplayName(mealType string) string
	GetDayDisplayName(day string) string

	StringSliceContains(slice []string, item string) bool
	StringSliceUnique(slice []string) []string
	StringSliceRemove(slice []string, item string) []string
	IntSliceContains(slice []int, item int) bool
	IntSliceUnique(slice []int) []int

	Base64Encode(data []byte) string
	Base64Decode(encoded string) ([]byte, error)
	URLSafeBase64Encode(data []byte) string
	URLSafeBase64Decode(encoded string) ([]byte, error)

	GenerateSecureToken(length int) string
	HashString(input string) string
	GenerateCSRFToken(ctx context.Context, sessionID string) (string, error)
	ValidateCSRFToken(ctx context.Context, sessionID, token string) (bool, error)
	CleanupExpiredCSRFTokens(ctx context.Context) error

	GetFileExtension(filename string) string
	GetMimeType(filename string) string
	IsImageFile(filename string) bool
	GenerateUniqueFilename(originalName string) string
	SanitizeFilename(filename string) string
}

type Service struct {
	redisClient *redis.Client
	csrfSecret  string
}

func New() Interface {
	return &Service{
		csrfSecret: "default-csrf-secret-change-in-production",
	}
}

func NewWithRedis(redisClient *redis.Client, csrfSecret string) Interface {
	if csrfSecret == "" {
		csrfSecret = "default-csrf-secret-change-in-production"
	}
	return &Service{
		redisClient: redisClient,
		csrfSecret:  csrfSecret,
	}
}

func (s *Service) GenerateULID() string {
	return ulid.Make().String()
}

func (s *Service) GenerateULIDWithTime(t time.Time) string {
	entropy := ulid.Monotonic(rand.Reader, 0)
	return ulid.MustNew(ulid.Timestamp(t), entropy).String()
}

func (s *Service) GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		randomInt, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[randomInt.Int64()]
	}
	return string(b)
}

func (s *Service) GenerateNumericOTP(length int) string {
	const charset = "0123456789"
	b := make([]byte, length)
	for i := range b {
		randomInt, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[randomInt.Int64()]
	}
	return string(b)
}

func (s *Service) GenerateAlphanumericCode(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		randomInt, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[randomInt.Int64()]
	}
	return string(b)
}

func (s *Service) IsValidEmail(email string) bool {
	const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

func (s *Service) IsValidPhone(phone string) bool {
	const phoneRegex = `^\+?[1-9]\d{6,14}$`
	re := regexp.MustCompile(phoneRegex)
	return re.MatchString(phone)
}

func (s *Service) IsValidIndonesianPhone(phone string) bool {
	cleaned := regexp.MustCompile(`[^\d+]`).ReplaceAllString(phone, "")
	patterns := []string{
		`^\+628\d{8,11}$`,
		`^628\d{8,11}$`,
		`^08\d{8,11}$`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, cleaned); matched {
			return true
		}
	}
	return false
}

func (s *Service) IsStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`\d`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>?]`).MatchString(password)

	return hasUpper && hasLower && hasNumber && hasSpecial
}

func (s *Service) IsValidURL(url string) bool {
	const urlRegex = `^https?://[^\s/$.?#].[^\s]*$`
	re := regexp.MustCompile(urlRegex)
	return re.MatchString(url)
}

func (s *Service) Slugify(text string) string {
	text = strings.ToLower(text)
	re := regexp.MustCompile(`[^a-z0-9]+`)
	text = re.ReplaceAllString(text, "-")
	text = strings.Trim(text, "-")
	return text
}

func (s *Service) TruncateString(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	if maxLength <= 3 {
		return text[:maxLength]
	}
	return text[:maxLength-3] + "..."
}

func (s *Service) CapitalizeFirst(text string) string {
	if text == "" {
		return text
	}
	return strings.ToUpper(text[:1]) + strings.ToLower(text[1:])
}

func (s *Service) CapitalizeWords(text string) string {
	return strings.Title(strings.ToLower(text))
}

func (s *Service) RemoveSpaces(text string) string {
	return strings.ReplaceAll(text, " ", "")
}

func (s *Service) NormalizeWhitespace(text string) string {
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(text, " "))
}

func (s *Service) FormatPrice(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}

func (s *Service) FormatPriceIDR(amount float64) string {
	return fmt.Sprintf("Rp %.2f", amount)
}

func (s *Service) ParsePrice(priceStr string) (float64, error) {
	cleaned := strings.ReplaceAll(priceStr, "Rp", "")
	cleaned = strings.ReplaceAll(cleaned, ",", "")
	cleaned = strings.TrimSpace(cleaned)
	return strconv.ParseFloat(cleaned, 64)
}

func (s *Service) RoundToDecimal(value float64, places int) float64 {
	multiplier := math.Pow(10, float64(places))
	return math.Round(value*multiplier) / multiplier
}

func (s *Service) CalculatePercentage(part, total float64) float64 {
	if total == 0 {
		return 0
	}
	return (part / total) * 100
}

func (s *Service) FormatDateID(t time.Time) string {
	return t.Format("02 January 2006")
}

func (s *Service) FormatDateTimeID(t time.Time) string {
	return t.Format("02 January 2006, 15:04")
}

func (s *Service) ParseDateID(dateStr string) (time.Time, error) {
	layouts := []string{
		"02 January 2006",
		"2 January 2006",
		"02-01-2006",
		"2-1-2006",
		"02/01/2006",
		"2/1/2006",
		"2006-01-02",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

func (s *Service) GetStartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func (s *Service) GetEndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

func (s *Service) GetStartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := t.AddDate(0, 0, -(weekday - 1))
	return s.GetStartOfDay(monday)
}

func (s *Service) GetEndOfWeek(t time.Time) time.Time {
	startOfWeek := s.GetStartOfWeek(t)
	sunday := startOfWeek.AddDate(0, 0, 6)
	return s.GetEndOfDay(sunday)
}

func (s *Service) GetStartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func (s *Service) GetEndOfMonth(t time.Time) time.Time {
	return s.GetStartOfMonth(t).AddDate(0, 1, 0).Add(-time.Nanosecond)
}

func (s *Service) DaysBetween(start, end time.Time) int {
	duration := end.Sub(start)
	return int(duration.Hours() / 24)
}

func (s *Service) WeeksBetween(start, end time.Time) int {
	return s.DaysBetween(start, end) / 7
}

func (s *Service) MonthsBetween(start, end time.Time) int {
	months := 0
	current := start
	for current.Before(end) {
		current = current.AddDate(0, 1, 0)
		if current.Before(end) || current.Equal(end) {
			months++
		}
	}
	return months
}

func (s *Service) CalculateOffset(page, limit int) int {
	if page <= 0 {
		page = 1
	}
	return (page - 1) * limit
}

func (s *Service) CalculateTotalPages(totalItems, limit int) int {
	if limit <= 0 {
		return 0
	}
	return int(math.Ceil(float64(totalItems) / float64(limit)))
}

func (s *Service) ValidatePagination(page, limit int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}
	return page, limit
}

func (s *Service) CalculateSubscriptionPrice(planPrice float64, mealTypes []string, deliveryDays []string) float64 {
	return planPrice * float64(len(mealTypes)) * float64(len(deliveryDays)) * 4.3
}

func (s *Service) ValidateMealTypes(mealTypes []string) error {
	validMealTypes := []string{"breakfast", "lunch", "dinner"}
	if len(mealTypes) == 0 {
		return fmt.Errorf("at least one meal type must be selected")
	}
	for _, mealType := range mealTypes {
		if !s.StringSliceContains(validMealTypes, strings.ToLower(mealType)) {
			return fmt.Errorf("invalid meal type: %s", mealType)
		}
	}
	return nil
}

func (s *Service) ValidateDeliveryDays(deliveryDays []string) error {
	validDays := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
	if len(deliveryDays) == 0 {
		return fmt.Errorf("at least one delivery day must be selected")
	}
	for _, day := range deliveryDays {
		if !s.StringSliceContains(validDays, strings.ToLower(day)) {
			return fmt.Errorf("invalid delivery day: %s", day)
		}
	}
	return nil
}

func (s *Service) GetMealTypeDisplayName(mealType string) string {
	mealTypeMap := map[string]string{
		"breakfast": "Breakfast",
		"lunch":     "Lunch",
		"dinner":    "Dinner",
	}
	if displayName, exists := mealTypeMap[strings.ToLower(mealType)]; exists {
		return displayName
	}
	return s.CapitalizeFirst(mealType)
}

func (s *Service) GetDayDisplayName(day string) string {
	dayMap := map[string]string{
		"monday":    "Monday",
		"tuesday":   "Tuesday",
		"wednesday": "Wednesday",
		"thursday":  "Thursday",
		"friday":    "Friday",
		"saturday":  "Saturday",
		"sunday":    "Sunday",
	}
	if displayName, exists := dayMap[strings.ToLower(day)]; exists {
		return displayName
	}
	return s.CapitalizeFirst(day)
}

func (s *Service) StringSliceContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (s *Service) StringSliceUnique(slice []string) []string {
	keys := make(map[string]bool)
	var result []string
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	return result
}

func (s *Service) StringSliceRemove(slice []string, item string) []string {
	var result []string
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

func (s *Service) IntSliceContains(slice []int, item int) bool {
	for _, i := range slice {
		if i == item {
			return true
		}
	}
	return false
}

func (s *Service) IntSliceUnique(slice []int) []int {
	keys := make(map[int]bool)
	var result []int
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	return result
}

func (s *Service) Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func (s *Service) Base64Decode(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

func (s *Service) URLSafeBase64Encode(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}

func (s *Service) URLSafeBase64Decode(encoded string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(encoded)
}

func (s *Service) GenerateSecureToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (s *Service) HashString(input string) string {
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", hash)
}

func (s *Service) GenerateCSRFToken(ctx context.Context, sessionID string) (string, error) {
	if sessionID == "" {
		return "", fmt.Errorf("session ID is required")
	}

	tokenData := make([]byte, 32)
	if _, err := rand.Read(tokenData); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	timestamp := time.Now().Unix()

	payload := fmt.Sprintf("%s:%d:%s", sessionID, timestamp, base64.URLEncoding.EncodeToString(tokenData))

	h := hmac.New(sha256.New, []byte(s.csrfSecret))
	h.Write([]byte(payload))
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))

	token := fmt.Sprintf("%s.%s", base64.URLEncoding.EncodeToString([]byte(payload)), signature)

	if s.redisClient != nil {
		csrfKey := fmt.Sprintf("csrf_token:%s:%s", sessionID, s.HashString(token))
		err := s.redisClient.Set(ctx, csrfKey, timestamp, 2*time.Hour).Err()
		if err != nil {
			return "", fmt.Errorf("failed to store CSRF token in Redis: %w", err)
		}
	}

	return token, nil
}

func (s *Service) ValidateCSRFToken(ctx context.Context, sessionID, token string) (bool, error) {
	if sessionID == "" || token == "" {
		return false, nil
	}

	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false, nil
	}

	payloadBytes, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return false, nil
	}

	providedSignature := parts[1]
	payload := string(payloadBytes)

	h := hmac.New(sha256.New, []byte(s.csrfSecret))
	h.Write([]byte(payload))
	expectedSignature := base64.URLEncoding.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(providedSignature), []byte(expectedSignature)) {
		return false, nil
	}

	payloadParts := strings.Split(payload, ":")
	if len(payloadParts) != 3 {
		return false, nil
	}

	tokenSessionID := payloadParts[0]
	timestampStr := payloadParts[1]

	if tokenSessionID != sessionID {
		return false, nil
	}

	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return false, nil
	}

	if time.Now().Unix()-timestamp > 2*60*60 {
		return false, nil
	}

	if s.redisClient != nil {
		csrfKey := fmt.Sprintf("csrf_token:%s:%s", sessionID, s.HashString(token))
		exists, err := s.redisClient.Exists(ctx, csrfKey).Result()
		if err != nil {

			return true, nil
		}

		if exists == 0 {
			return false, nil
		}

		s.redisClient.Del(ctx, csrfKey)
	}

	return true, nil
}

func (s *Service) CleanupExpiredCSRFTokens(ctx context.Context) error {
	if s.redisClient == nil {
		return nil
	}

	keys, err := s.redisClient.Keys(ctx, "csrf_token:*").Result()
	if err != nil {
		return fmt.Errorf("failed to get CSRF token keys: %w", err)
	}

	expiredKeys := make([]string, 0)
	for _, key := range keys {
		ttl, err := s.redisClient.TTL(ctx, key).Result()
		if err != nil {
			continue
		}

		if ttl == -1 || ttl == -2 {
			expiredKeys = append(expiredKeys, key)
		}
	}

	if len(expiredKeys) > 0 {
		err = s.redisClient.Del(ctx, expiredKeys...).Err()
		if err != nil {
			return fmt.Errorf("failed to cleanup expired CSRF tokens: %w", err)
		}
	}

	return nil
}

func (s *Service) GetFileExtension(filename string) string {
	return strings.ToLower(filepath.Ext(filename))
}

func (s *Service) GetMimeType(filename string) string {
	ext := s.GetFileExtension(filename)
	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".pdf":  "application/pdf",
		".txt":  "text/plain",
		".html": "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".json": "application/json",
	}

	if mimeType, exists := mimeTypes[ext]; exists {
		return mimeType
	}
	return "application/octet-stream"
}

func (s *Service) IsImageFile(filename string) bool {
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg"}
	ext := s.GetFileExtension(filename)
	return s.StringSliceContains(imageExtensions, ext)
}

func (s *Service) GenerateUniqueFilename(originalName string) string {
	ext := s.GetFileExtension(originalName)
	nameWithoutExt := strings.TrimSuffix(originalName, ext)
	timestamp := time.Now().Unix()
	randomStr := s.GenerateRandomString(8)
	return fmt.Sprintf("%s_%d_%s%s", s.SanitizeFilename(nameWithoutExt), timestamp, randomStr, ext)
}

func (s *Service) SanitizeFilename(filename string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	sanitized := reg.ReplaceAllString(filename, "_")
	reg = regexp.MustCompile(`_{2,}`)
	sanitized = reg.ReplaceAllString(sanitized, "_")
	return strings.Trim(sanitized, "_")
}
