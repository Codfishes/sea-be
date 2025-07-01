package logger

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	formatter "github.com/antonfisher/nested-logrus-formatter"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	instance *Logger
	once     sync.Once
)

type Logger struct {
	*logrus.Logger
	config *Config
}

type Config struct {
	Level       string
	Format      string
	Output      string
	FilePath    string
	MaxSize     int
	MaxAge      int
	MaxBackups  int
	Compress    bool
	ServiceName string
}

type Fields = logrus.Fields

type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
	PanicLevel LogLevel = "panic"
)

type LogFormat string

const (
	JSONFormat LogFormat = "json"
	TextFormat LogFormat = "text"
)

type LogOutput string

const (
	StdoutOutput LogOutput = "stdout"
	FileOutput   LogOutput = "file"
	BothOutput   LogOutput = "both"
)

func DefaultConfig() *Config {
	return &Config{
		Level:       string(InfoLevel),
		Format:      string(TextFormat),
		Output:      string(StdoutOutput),
		FilePath:    "./logs/app.log",
		MaxSize:     100,
		MaxAge:      7,
		MaxBackups:  3,
		Compress:    true,
		ServiceName: "sea-catering-backend",
	}
}

func LoadConfigFromEnv() *Config {
	config := DefaultConfig()

	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Level = level
	}
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		config.Format = format
	}
	if output := os.Getenv("LOG_OUTPUT"); output != "" {
		config.Output = output
	}
	if filePath := os.Getenv("LOG_FILE_PATH"); filePath != "" {
		config.FilePath = filePath
	}
	if serviceName := os.Getenv("APP_NAME"); serviceName != "" {
		config.ServiceName = serviceName
	}

	return config
}

func New(config *Config) *Logger {
	if config == nil {
		config = DefaultConfig()
	}

	logger := logrus.New()

	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	switch LogFormat(config.Format) {
	case JSONFormat:
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "function",
				logrus.FieldKeyFile:  "file",
			},
		})
	default:
		logger.SetFormatter(&formatter.Formatter{
			NoColors:        false,
			TimestampFormat: "2006-01-02 15:04:05",
			HideKeys:        false,
			CallerFirst:     true,
			CustomCallerFormatter: func(f *runtime.Frame) string {
				s := strings.Split(f.Function, ".")
				funcName := s[len(s)-1]
				return fmt.Sprintf(" \x1b[%dm[%s:%d][%s()]", 34, path.Base(f.File), f.Line, funcName)
			},
		})
	}

	var writers []io.Writer
	switch LogOutput(config.Output) {
	case FileOutput:
		writers = []io.Writer{createFileWriter(config)}
	case BothOutput:
		writers = []io.Writer{os.Stdout, createFileWriter(config)}
	default:
		writers = []io.Writer{os.Stdout}
	}

	logger.SetOutput(io.MultiWriter(writers...))
	logger.SetReportCaller(true)

	return &Logger{
		Logger: logger,
		config: config,
	}
}

func GetInstance() *Logger {
	once.Do(func() {
		config := LoadConfigFromEnv()
		instance = New(config)
	})
	return instance
}

func createFileWriter(config *Config) io.Writer {

	if err := os.MkdirAll(path.Dir(config.FilePath), 0755); err != nil {
		logrus.Fatalf("Failed to create log directory: %v", err)
	}

	return &lumberjack.Logger{
		Filename:   config.FilePath,
		MaxSize:    config.MaxSize,
		MaxAge:     config.MaxAge,
		MaxBackups: config.MaxBackups,
		LocalTime:  true,
		Compress:   config.Compress,
	}
}

func (l *Logger) WithRequestID(requestID string) *logrus.Entry {
	if requestID == "" {
		requestID = generateRequestID()
	}
	return l.WithField("request_id", requestID)
}

func (l *Logger) WithUserID(userID string) *logrus.Entry {
	return l.WithField("user_id", userID)
}

func (l *Logger) WithService(service string) *logrus.Entry {
	return l.WithField("service", service)
}

func (l *Logger) WithModule(module string) *logrus.Entry {
	return l.WithField("module", module)
}

func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}

func (l *Logger) WithFields(fields Fields) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields(fields))
}

func (l *Logger) Debug(msg string, fields ...Fields) {
	entry := l.Logger.WithFields(mergeFields(fields...))
	entry.Debug(msg)
}

func (l *Logger) Info(msg string, fields ...Fields) {
	entry := l.Logger.WithFields(mergeFields(fields...))
	entry.Info(msg)
}

func (l *Logger) Warn(msg string, fields ...Fields) {
	entry := l.Logger.WithFields(mergeFields(fields...))
	entry.Warn(msg)
}

func (l *Logger) Error(msg string, fields ...Fields) {
	entry := l.Logger.WithFields(mergeFields(fields...))
	entry.Error(msg)
}

func (l *Logger) Fatal(msg string, fields ...Fields) {
	entry := l.Logger.WithFields(mergeFields(fields...))
	entry.Fatal(msg)
}

func (l *Logger) Panic(msg string, fields ...Fields) {
	entry := l.Logger.WithFields(mergeFields(fields...))
	entry.Panic(msg)
}

func (l *Logger) LogRequest(method, path, ip string, duration time.Duration, status int, fields ...Fields) {
	baseFields := Fields{
		"method":   method,
		"path":     path,
		"ip":       ip,
		"duration": duration.String(),
		"status":   status,
	}

	allFields := mergeFields(append([]Fields{baseFields}, fields...)...)
	entry := l.Logger.WithFields(allFields)

	msg := fmt.Sprintf("%s %s - %d", method, path, status)

	switch {
	case status >= 500:
		entry.Error(msg)
	case status >= 400:
		entry.Warn(msg)
	default:
		entry.Info(msg)
	}
}

func (l *Logger) LogDatabaseQuery(query string, duration time.Duration, err error, fields ...Fields) {
	baseFields := Fields{
		"query":    query,
		"duration": duration.String(),
	}

	if err != nil {
		baseFields["error"] = err.Error()
	}

	allFields := mergeFields(append([]Fields{baseFields}, fields...)...)
	entry := l.Logger.WithFields(allFields)

	if err != nil {
		entry.Error("Database query failed")
	} else {
		entry.Debug("Database query executed")
	}
}

func (l *Logger) LogAPICall(url, method string, duration time.Duration, status int, err error, fields ...Fields) {
	baseFields := Fields{
		"url":      url,
		"method":   method,
		"duration": duration.String(),
		"status":   status,
	}

	if err != nil {
		baseFields["error"] = err.Error()
	}

	allFields := mergeFields(append([]Fields{baseFields}, fields...)...)
	entry := l.Logger.WithFields(allFields)

	msg := fmt.Sprintf("API call: %s %s - %d", method, url, status)

	if err != nil || status >= 400 {
		entry.Error(msg)
	} else {
		entry.Info(msg)
	}
}

func (l *Logger) SetLevel(level LogLevel) {
	logrusLevel, err := logrus.ParseLevel(string(level))
	if err != nil {
		l.Warn("Invalid log level, keeping current level", Fields{"level": level})
		return
	}
	l.Logger.SetLevel(logrusLevel)
	l.Info("Log level changed", Fields{"new_level": level})
}

func (l *Logger) GetLevel() LogLevel {
	return LogLevel(l.Logger.GetLevel().String())
}

func (l *Logger) IsDebugEnabled() bool {
	return l.Logger.IsLevelEnabled(logrus.DebugLevel)
}

func (l *Logger) IsInfoEnabled() bool {
	return l.Logger.IsLevelEnabled(logrus.InfoLevel)
}

func generateRequestID() string {
	return uuid.New().String()
}

func mergeFields(fieldMaps ...Fields) logrus.Fields {
	result := make(logrus.Fields)
	for _, fields := range fieldMaps {
		if fields != nil {
			for k, v := range fields {
				result[k] = v
			}
		}
	}
	return result
}

func Debug(msg string, fields ...Fields) {
	GetInstance().Debug(msg, fields...)
}

func Info(msg string, fields ...Fields) {
	GetInstance().Info(msg, fields...)
}

func Warn(msg string, fields ...Fields) {
	GetInstance().Warn(msg, fields...)
}

func Error(msg string, fields ...Fields) {
	GetInstance().Error(msg, fields...)
}

func Fatal(msg string, fields ...Fields) {
	GetInstance().Fatal(msg, fields...)
}

func Panic(msg string, fields ...Fields) {
	GetInstance().Panic(msg, fields...)
}

func WithRequestID(requestID string) *logrus.Entry {
	return GetInstance().WithRequestID(requestID)
}

func WithUserID(userID string) *logrus.Entry {
	return GetInstance().WithUserID(userID)
}

func WithService(service string) *logrus.Entry {
	return GetInstance().WithService(service)
}

func WithModule(module string) *logrus.Entry {
	return GetInstance().WithModule(module)
}

func WithError(err error) *logrus.Entry {
	return GetInstance().WithError(err)
}

func WithFields(fields Fields) *logrus.Entry {
	return GetInstance().WithFields(fields)
}
