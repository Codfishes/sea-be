package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Interface interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	GetStruct(ctx context.Context, key string, dest interface{}) error
	Del(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (int64, error)
	TTL(ctx context.Context, key string) (time.Duration, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error

	HSet(ctx context.Context, key string, values ...interface{}) error
	HGet(ctx context.Context, key, field string) (string, error)
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HDel(ctx context.Context, key string, fields ...string) error

	LPush(ctx context.Context, key string, values ...interface{}) error
	RPush(ctx context.Context, key string, values ...interface{}) error
	LPop(ctx context.Context, key string) (string, error)
	RPop(ctx context.Context, key string) (string, error)
	LLen(ctx context.Context, key string) (int64, error)
	LRange(ctx context.Context, key string, start, stop int64) ([]string, error)

	SAdd(ctx context.Context, key string, members ...interface{}) error
	SRem(ctx context.Context, key string, members ...interface{}) error
	SMembers(ctx context.Context, key string) ([]string, error)
	SIsMember(ctx context.Context, key string, member interface{}) (bool, error)

	SetCache(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	GetCache(ctx context.Context, key string, dest interface{}) error
	DeleteCache(ctx context.Context, pattern string) error

	SetSession(ctx context.Context, sessionID string, data interface{}, expiration time.Duration) error
	GetSession(ctx context.Context, sessionID string, dest interface{}) error
	DeleteSession(ctx context.Context, sessionID string) error

	SetOTP(ctx context.Context, identifier string, otp string, expiration time.Duration) error
	GetOTP(ctx context.Context, identifier string) (string, error)
	DeleteOTP(ctx context.Context, identifier string) error
	VerifyOTP(ctx context.Context, identifier, otp string) (bool, error)

	IsRateLimited(ctx context.Context, key string, limit int64, window time.Duration) (bool, error)
	IncrementRate(ctx context.Context, key string, window time.Duration) (int64, error)

	Publish(ctx context.Context, channel string, message interface{}) error
	Subscribe(ctx context.Context, channels ...string) *redis.PubSub

	Ping(ctx context.Context) error
	Close() error
	GetClient() *redis.Client
}

type Service struct {
	client *redis.Client
	config *Config
}

type Config struct {
	Host         string
	Port         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolTimeout  time.Duration
	IdleTimeout  time.Duration
}

func LoadConfig() *Config {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}

	password := os.Getenv("REDIS_PASSWORD")

	db := 0
	if envDB := os.Getenv("REDIS_DB"); envDB != "" {
		if parsed, err := strconv.Atoi(envDB); err == nil {
			db = parsed
		}
	}

	poolSize := 10
	if envPoolSize := os.Getenv("REDIS_POOL_SIZE"); envPoolSize != "" {
		if parsed, err := strconv.Atoi(envPoolSize); err == nil {
			poolSize = parsed
		}
	}

	return &Config{
		Host:         host,
		Port:         port,
		Password:     password,
		DB:           db,
		PoolSize:     poolSize,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
		IdleTimeout:  5 * time.Minute,
	}
}

func New() (Interface, error) {
	config := LoadConfig()
	return NewWithConfig(config)
}

func NewWithConfig(config *Config) (Interface, error) {
	if config == nil {
		config = LoadConfig()
	}

	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		PoolTimeout:  config.PoolTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Service{
		client: client,
		config: config,
	}, nil
}

func (s *Service) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var val interface{}

	switch v := value.(type) {
	case string, int, int64, float64, bool:
		val = v
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		val = string(data)
	}

	return s.client.Set(ctx, key, val, expiration).Err()
}

func (s *Service) Get(ctx context.Context, key string) (string, error) {
	val, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key not found")
		}
		return "", fmt.Errorf("failed to get key: %w", err)
	}
	return val, nil
}

func (s *Service) GetStruct(ctx context.Context, key string, dest interface{}) error {
	val, err := s.Get(ctx, key)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

func (s *Service) Del(ctx context.Context, keys ...string) error {
	return s.client.Del(ctx, keys...).Err()
}

func (s *Service) Exists(ctx context.Context, keys ...string) (int64, error) {
	return s.client.Exists(ctx, keys...).Result()
}

func (s *Service) TTL(ctx context.Context, key string) (time.Duration, error) {
	return s.client.TTL(ctx, key).Result()
}

func (s *Service) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return s.client.Expire(ctx, key, expiration).Err()
}

func (s *Service) HSet(ctx context.Context, key string, values ...interface{}) error {
	return s.client.HSet(ctx, key, values...).Err()
}

func (s *Service) HGet(ctx context.Context, key, field string) (string, error) {
	return s.client.HGet(ctx, key, field).Result()
}

func (s *Service) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return s.client.HGetAll(ctx, key).Result()
}

func (s *Service) HDel(ctx context.Context, key string, fields ...string) error {
	return s.client.HDel(ctx, key, fields...).Err()
}

func (s *Service) LPush(ctx context.Context, key string, values ...interface{}) error {
	return s.client.LPush(ctx, key, values...).Err()
}

func (s *Service) RPush(ctx context.Context, key string, values ...interface{}) error {
	return s.client.RPush(ctx, key, values...).Err()
}

func (s *Service) LPop(ctx context.Context, key string) (string, error) {
	return s.client.LPop(ctx, key).Result()
}

func (s *Service) RPop(ctx context.Context, key string) (string, error) {
	return s.client.RPop(ctx, key).Result()
}

func (s *Service) LLen(ctx context.Context, key string) (int64, error) {
	return s.client.LLen(ctx, key).Result()
}

func (s *Service) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return s.client.LRange(ctx, key, start, stop).Result()
}

func (s *Service) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return s.client.SAdd(ctx, key, members...).Err()
}

func (s *Service) SRem(ctx context.Context, key string, members ...interface{}) error {
	return s.client.SRem(ctx, key, members...).Err()
}

func (s *Service) SMembers(ctx context.Context, key string) ([]string, error) {
	return s.client.SMembers(ctx, key).Result()
}

func (s *Service) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return s.client.SIsMember(ctx, key, member).Result()
}

func (s *Service) SetCache(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return s.Set(ctx, fmt.Sprintf("cache:%s", key), value, expiration)
}

func (s *Service) GetCache(ctx context.Context, key string, dest interface{}) error {
	return s.GetStruct(ctx, fmt.Sprintf("cache:%s", key), dest)
}

func (s *Service) DeleteCache(ctx context.Context, pattern string) error {
	keys, err := s.client.Keys(ctx, fmt.Sprintf("cache:%s", pattern)).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return s.Del(ctx, keys...)
	}

	return nil
}

func (s *Service) SetSession(ctx context.Context, sessionID string, data interface{}, expiration time.Duration) error {
	return s.Set(ctx, fmt.Sprintf("session:%s", sessionID), data, expiration)
}

func (s *Service) GetSession(ctx context.Context, sessionID string, dest interface{}) error {
	return s.GetStruct(ctx, fmt.Sprintf("session:%s", sessionID), dest)
}

func (s *Service) DeleteSession(ctx context.Context, sessionID string) error {
	return s.Del(ctx, fmt.Sprintf("session:%s", sessionID))
}

func (s *Service) SetOTP(ctx context.Context, identifier string, otp string, expiration time.Duration) error {
	return s.Set(ctx, fmt.Sprintf("otp:%s", identifier), otp, expiration)
}

func (s *Service) GetOTP(ctx context.Context, identifier string) (string, error) {
	return s.Get(ctx, fmt.Sprintf("otp:%s", identifier))
}

func (s *Service) DeleteOTP(ctx context.Context, identifier string) error {
	return s.Del(ctx, fmt.Sprintf("otp:%s", identifier))
}

func (s *Service) VerifyOTP(ctx context.Context, identifier, otp string) (bool, error) {
	storedOTP, err := s.GetOTP(ctx, identifier)
	if err != nil {
		return false, err
	}

	if storedOTP == otp {

		s.DeleteOTP(ctx, identifier)
		return true, nil
	}

	return false, nil
}

func (s *Service) IsRateLimited(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	current, err := s.IncrementRate(ctx, key, window)
	if err != nil {
		return false, err
	}

	return current > limit, nil
}

func (s *Service) IncrementRate(ctx context.Context, key string, window time.Duration) (int64, error) {
	rateLimitKey := fmt.Sprintf("rate_limit:%s", key)

	pipe := s.client.Pipeline()
	incr := pipe.Incr(ctx, rateLimitKey)
	pipe.Expire(ctx, rateLimitKey, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	return incr.Val(), nil
}

func (s *Service) Publish(ctx context.Context, channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return s.client.Publish(ctx, channel, string(data)).Err()
}

func (s *Service) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return s.client.Subscribe(ctx, channels...)
}

func (s *Service) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

func (s *Service) Close() error {
	return s.client.Close()
}

func (s *Service) GetClient() *redis.Client {
	return s.client
}

func GenerateKey(prefix string, parts ...string) string {
	key := prefix
	for _, part := range parts {
		key += ":" + part
	}
	return key
}

type BatchOperation struct {
	Operation string
	Key       string
	Value     interface{}
	TTL       time.Duration
}

func (s *Service) ExecuteBatch(ctx context.Context, operations []BatchOperation) error {
	pipe := s.client.Pipeline()

	for _, op := range operations {
		switch op.Operation {
		case "SET":
			pipe.Set(ctx, op.Key, op.Value, op.TTL)
		case "GET":
			pipe.Get(ctx, op.Key)
		case "DEL":
			pipe.Del(ctx, op.Key)
		case "HSET":
			pipe.HSet(ctx, op.Key, op.Value)
		case "HDEL":
			if values, ok := op.Value.([]string); ok {
				pipe.HDel(ctx, op.Key, values...)
			}
		}
	}

	_, err := pipe.Exec(ctx)
	return err
}
