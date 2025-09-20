package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/rs/zerolog"
	zerologlog "github.com/rs/zerolog/log"
	devtrace "github.com/skulidropek/gotrace"
)

const defaultLogFile = "logs/app.log"

var (
	initOnce       sync.Once
	devTraceOutput zerolog.LevelWriter
)

// Init configures zerolog and installs DevTrace stack logging once per process.
func Init() {
	initOnce.Do(func() {
		configureZerolog()
		configureDevTrace()
	})
}

func init() {
	Init()
}

func configureZerolog() {
	level := zerolog.DebugLevel
	if env := strings.TrimSpace(os.Getenv("LOG_LEVEL")); env != "" {
		if parsed, err := zerolog.ParseLevel(strings.ToLower(env)); err == nil {
			level = parsed
		}
	}

	zerolog.SetGlobalLevel(level)

	writers := buildWriters()
	base := zerolog.MultiLevelWriter(writers...)
	devTraceOutput = base
	zerologlog.Logger = zerolog.New(devTraceLevelWriter{base: base}).With().Timestamp().Logger()
}

func buildWriters() []io.Writer {
	writers := make([]io.Writer, 0, 2)

	logPath := strings.TrimSpace(os.Getenv("LOG_FILE"))
	if logPath == "" {
		logPath = defaultLogFile
	}
	if logPath != "" {
		if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "logging: failed to create log directory for %s: %v\n", logPath, err)
		} else {
			file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "logging: failed to open log file %s: %v\n", logPath, err)
			} else {
				writers = append(writers, file)
			}
		}
	}

	if parseBoolEnv("LOG_CONSOLE", true) {
		writers = append(writers, zerolog.ConsoleWriter{Out: os.Stderr})
	}

	if len(writers) == 0 {
		writers = append(writers, zerolog.ConsoleWriter{Out: os.Stderr})
	}

	return writers
}

func configureDevTrace() {
	cfg := devtrace.DefaultConfig

	cfg.Enabled = parseEnabledEnv(cfg.Enabled)
	cfg.StackLimit = parseIntEnv("DEVTRACE_STACK_LIMIT", cfg.StackLimit)
	cfg.ShowSnippet = parseIntEnv("DEVTRACE_SHOW_SNIPPET", 2)
	cfg.AppPattern = parseStringEnv("DEVTRACE_APP_PATTERN", "github.com/Swarmind/libagent")
	cfg.DebugLevel = parseIntEnv("DEVTRACE_DEBUG_LEVEL", 1)

	devtrace.SetConfig(cfg)

	opts := devtrace.DefaultStackLoggerOptions
	opts.Prefix = "CALL STACK"
	opts.Skip = parseIntEnv("DEVTRACE_STACK_SKIP", opts.Skip)
	opts.Limit = cfg.StackLimit
	opts.ShowSnippet = cfg.ShowSnippet
	opts.OnlyApp = parseBoolEnv("DEVTRACE_ONLY_APP", opts.OnlyApp)
	opts.PreferApp = parseBoolEnv("DEVTRACE_PREFER_APP", true)
	opts.AppPattern = cfg.AppPattern
	opts.Ascending = parseBoolEnv("DEVTRACE_ASCENDING", opts.Ascending)

	var logger devtrace.Logger
	if devTraceOutput != nil {
		logger = &devTraceZerologLogger{writer: devTraceOutput}
	} else {
		fallback := zerolog.LevelWriterAdapter{Writer: zerolog.ConsoleWriter{Out: os.Stderr}}
		logger = &devTraceZerologLogger{writer: fallback}
	}
	devtrace.SetLogger(logger)
	devtrace.InstallStackLogger(&opts)
	devtrace.GlobalEnhancedLogger.SetLogger(logger)
	devtrace.RedirectStandardLogger()
}

func parseEnabledEnv(defaultVal bool) bool {
	env := strings.TrimSpace(os.Getenv("DEVTRACE_ENABLED"))
	if env == "" {
		if !defaultVal {
			return true
		}
		return defaultVal
	}
	return parseBool(env)
}

func parseBoolEnv(key string, defaultVal bool) bool {
	env := strings.TrimSpace(os.Getenv(key))
	if env == "" {
		return defaultVal
	}
	return parseBool(env)
}

func parseBool(value string) bool {
	switch strings.ToLower(value) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return false
	}
}

func parseIntEnv(key string, defaultVal int) int {
	env := strings.TrimSpace(os.Getenv(key))
	if env == "" {
		return defaultVal
	}
	if v, err := strconv.Atoi(env); err == nil {
		return v
	}
	return defaultVal
}

func parseStringEnv(key, defaultVal string) string {
	env := strings.TrimSpace(os.Getenv(key))
	if env == "" {
		return defaultVal
	}
	return env
}

type devTraceLevelWriter struct {
	base zerolog.LevelWriter
}

func (w devTraceLevelWriter) Write(p []byte) (int, error) {
	return w.base.Write(p)
}

func (w devTraceLevelWriter) WriteLevel(level zerolog.Level, p []byte) (int, error) {
	if devtrace.IsEnabled() {
		logLevel := mapZerologLevel(level)
		message := extractLogMessage(p)
		devtrace.GlobalEnhancedLogger.LogWithStack(context.Background(), logLevel, message)
	}
	return w.base.WriteLevel(level, p)
}

func mapZerologLevel(level zerolog.Level) string {
	switch level {
	case zerolog.DebugLevel:
		return "DEBUG"
	case zerolog.InfoLevel, zerolog.NoLevel:
		return "INFO"
	case zerolog.WarnLevel:
		return "WARN"
	case zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel:
		return "ERROR"
	default:
		return "INFO"
	}
}

func extractLogMessage(p []byte) string {
	var payload map[string]interface{}
	if err := json.Unmarshal(p, &payload); err == nil {
		message := fmt.Sprint(payload["message"])
		if message == "" || message == "<nil>" {
			message = strings.TrimSpace(string(p))
		}
		if errVal, ok := payload["error"]; ok && errVal != nil {
			return fmt.Sprintf("%s | error=%v", message, errVal)
		}
		return message
	}
	return strings.TrimSpace(string(p))
}

type devTraceZerologLogger struct {
	writer zerolog.LevelWriter
}

func (l *devTraceZerologLogger) Log(level string, msg string, args ...interface{}) {
	zerologLevel := parseDevTraceLevel(level)
	formatted := fmt.Sprintf("[DEVTRACE-%s] %s", level, fmt.Sprintf(msg, args...))
	_, _ = l.writer.WriteLevel(zerologLevel, append([]byte(formatted), '\n'))
}

func (l *devTraceZerologLogger) Debug(msg string, args ...interface{}) {
	l.Log("DEBUG", msg, args...)
}

func (l *devTraceZerologLogger) Info(msg string, args ...interface{}) {
	l.Log("INFO", msg, args...)
}

func (l *devTraceZerologLogger) Warn(msg string, args ...interface{}) {
	l.Log("WARN", msg, args...)
}

func (l *devTraceZerologLogger) Error(msg string, args ...interface{}) {
	l.Log("ERROR", msg, args...)
}

func parseDevTraceLevel(level string) zerolog.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return zerolog.DebugLevel
	case "INFO":
		return zerolog.InfoLevel
	case "WARN":
		return zerolog.WarnLevel
	case "ERROR":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}
