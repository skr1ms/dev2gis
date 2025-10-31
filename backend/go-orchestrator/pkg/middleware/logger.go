package middleware

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/skr1ms/dev2gis/config"
)

type Logger struct {
	logger zerolog.Logger
	config *config.Config
}

type LoggerInterface interface {
	Debug(msg string, fields ...map[string]interface{})
	Info(msg string, fields ...map[string]interface{})
	Warn(msg string, fields ...map[string]interface{})
	Error(msg string, err error, fields ...map[string]interface{})
	Fatal(msg string, err error, fields ...map[string]interface{})
	WithContext(ctx context.Context) *Logger
	Middleware() gin.HandlerFunc
}

func NewLogger(cfg *config.Config) LoggerInterface {
	var writer io.Writer = os.Stdout

	if cfg.Server.Environment != "production" {
		writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		}
	}

	zerolog.SetGlobalLevel(cfg.Server.LogLevel)

	logger := zerolog.New(writer).With().
		Timestamp().
		Str("service", "dev2gis-gateway").
		Str("version", "1.0.0").
		Str("environment", cfg.Server.Environment).
		Logger()

	return &Logger{
		logger: logger,
		config: cfg,
	}
}

func (l *Logger) Debug(msg string, fields ...map[string]interface{}) {
	event := l.logger.Debug()
	l.addFields(event, fields...)
	event.Msg(msg)
}

func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	event := l.logger.Info()
	l.addFields(event, fields...)
	event.Msg(msg)
}

func (l *Logger) Warn(msg string, fields ...map[string]interface{}) {
	event := l.logger.Warn()
	l.addFields(event, fields...)
	event.Msg(msg)
}

func (l *Logger) Error(msg string, err error, fields ...map[string]interface{}) {
	event := l.logger.Error()
	if err != nil {
		event.Err(err)
	}
	l.addFields(event, fields...)
	event.Msg(msg)
}

func (l *Logger) Fatal(msg string, err error, fields ...map[string]interface{}) {
	event := l.logger.Fatal()
	if err != nil {
		event.Err(err)
	}
	l.addFields(event, fields...)
	event.Msg(msg)
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
	return &Logger{
		logger: l.logger.With().Logger(),
		config: l.config,
	}
}

func (l *Logger) addFields(event *zerolog.Event, fields ...map[string]interface{}) {
	for _, fieldMap := range fields {
		for key, value := range fieldMap {
			event.Interface(key, value)
		}
	}
}

func (l *Logger) Middleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		var statusColor, methodColor, resetColor string
		if l.config.Server.Environment != "production" {
			statusColor = param.StatusCodeColor()
			methodColor = param.MethodColor()
			resetColor = param.ResetColor()
		}

		if param.Latency > time.Minute {
			param.Latency = param.Latency - param.Latency%time.Second
		}

		logData := map[string]interface{}{
			"method":        param.Method,
			"path":          param.Path,
			"status":        param.StatusCode,
			"latency":       param.Latency.String(),
			"client_ip":     param.ClientIP,
			"user_agent":    param.Request.UserAgent(),
			"response_size": param.BodySize,
		}

		if param.ErrorMessage != "" {
			logData["error"] = param.ErrorMessage
			l.Error("HTTP Request", nil, logData)
		} else {
			l.Info("HTTP Request", logData)
		}

		if l.config.Server.Environment != "production" {
			return fmt.Sprintf("%v |%s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s",
				param.TimeStamp.Format("2006/01/02 - 15:04:05"),
				statusColor, param.StatusCode, resetColor,
				param.Latency,
				param.ClientIP,
				methodColor, param.Method, resetColor,
				param.Path,
				param.ErrorMessage,
			)
		}

		return ""
	})
}

var GlobalLogger LoggerInterface

func SetGlobalLogger(logger LoggerInterface) {
	GlobalLogger = logger
	if l, ok := logger.(*Logger); ok {
		log.Logger = l.logger
	}
}

func Debug(msg string, fields ...map[string]interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Debug(msg, fields...)
	}
}

func Info(msg string, fields ...map[string]interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Info(msg, fields...)
	}
}

func Warn(msg string, fields ...map[string]interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Warn(msg, fields...)
	}
}

func Error(msg string, err error, fields ...map[string]interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Error(msg, err, fields...)
	}
}

func Fatal(msg string, err error, fields ...map[string]interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Fatal(msg, err, fields...)
	}
}
