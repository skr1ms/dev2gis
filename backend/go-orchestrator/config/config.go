package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

type Config struct {
	Server   ServerConfig
	DB       DBConfig
	Minio    MinioConfig
	Auth     AuthConfig
	RabbitMQ RabbitMQConfig
}

type ServerConfig struct {
	LogLevel    zerolog.Level
	Environment string
	BackendPort string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type MinioConfig struct {
	Endpoint                 string
	AccessKeyID              string
	SecretAccessKey          string
	UseSSL                   bool
	Region                   string
	UAVDataBucketName        string
	UAVModelsBucketName      string
	UAVPhotoplanesBucketName string
	PublicURL                string
}

type AuthConfig struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	JWTAccessTokenTTL  time.Duration
	JWTRefreshTokenTTL time.Duration
}

type RabbitMQConfig struct {
	URL               string
	QueueName         string
	Exchange          string
	RoutingKey        string
	PrefetchCount     int
	ConnectionTimeout time.Duration
	ReconnectDelay    time.Duration
	PublishRetries    int
	DurableQueue      bool
	AutoDelete        bool
}

func NewConfig() (*Config, error) {
	envPath := os.Getenv("ENV_FILE")
	if envPath == "" {
		env := getEnvironment()
		if env == "production" || env == "prod" {
			envPath = "../../.env.prod"
		} else {
			envPath = "../../.env.dev"
		}
	}
	err := godotenv.Load(envPath)
	if err != nil {
		log.Printf("Warning: .env file not found: %s", envPath)
	}

	environment := getEnvironment()
	if err := validateEnvironment(environment); err != nil {
		return nil, fmt.Errorf("invalid environment: %w", err)
	}

	logLevel := parseLogLevel(os.Getenv("LOG_LEVEL"))
	backendPort := getEnvOrDefault("PORT", "8080")

	return &Config{
		Server: ServerConfig{
			LogLevel:    logLevel,
			Environment: environment,
			BackendPort: backendPort,
		},
		DB: DBConfig{
			Host:     os.Getenv("POSTGRES_HOST"),
			Port:     os.Getenv("POSTGRES_PORT"),
			User:     os.Getenv("POSTGRES_USER"),
			Password: os.Getenv("POSTGRES_PASSWORD"),
			DBName:   os.Getenv("POSTGRES_DB"),
		},
		Minio: MinioConfig{
			Endpoint:                 os.Getenv("MINIO_ENDPOINT"),
			AccessKeyID:              os.Getenv("MINIO_ROOT_USER"),
			SecretAccessKey:          os.Getenv("MINIO_ROOT_PASSWORD"),
			UseSSL:                   false,
			Region:                   os.Getenv("MINIO_REGION"),
			UAVDataBucketName:        os.Getenv("MINIO_UAV_DATA_BUCKET"),
			UAVModelsBucketName:      os.Getenv("MINIO_UAV_MODELS_BUCKET"),
			UAVPhotoplanesBucketName: os.Getenv("MINIO_UAV_PHOTOPLANES_BUCKET"),
			PublicURL:                os.Getenv("MINIO_PUBLIC_URL"),
		},
		Auth: AuthConfig{
			AccessTokenSecret:  os.Getenv("ACCESS_TOKEN_SECRET"),
			RefreshTokenSecret: os.Getenv("REFRESH_TOKEN_SECRET"),
			JWTAccessTokenTTL:  parseDuration(os.Getenv("JWT_ACCESS_TTL")),
			JWTRefreshTokenTTL: parseDuration(os.Getenv("JWT_REFRESH_TTL")),
		},
		RabbitMQ: RabbitMQConfig{
			URL:               getEnvOrDefault("RABBITMQ_URL", "amqp://admin:admin@rabbitmq:5672/"),
			QueueName:         getEnvOrDefault("RABBITMQ_QUEUE_NAME", "heightmap.tasks"),
			Exchange:          getEnvOrDefault("RABBITMQ_EXCHANGE", ""),
			RoutingKey:        getEnvOrDefault("RABBITMQ_ROUTING_KEY", "heightmap.tasks"),
			PrefetchCount:     parseInt(getEnvOrDefault("RABBITMQ_PREFETCH_COUNT", "1")),
			ConnectionTimeout: parseDuration(getEnvOrDefault("RABBITMQ_CONNECTION_TIMEOUT", "30s")),
			ReconnectDelay:    parseDuration(getEnvOrDefault("RABBITMQ_RECONNECT_DELAY", "5s")),
			PublishRetries:    parseInt(getEnvOrDefault("RABBITMQ_PUBLISH_RETRIES", "3")),
			DurableQueue:      parseBool(getEnvOrDefault("RABBITMQ_DURABLE_QUEUE", "true")),
			AutoDelete:        parseBool(getEnvOrDefault("RABBITMQ_AUTO_DELETE", "false")),
		},
	}, nil
}

func parseDuration(duration string) time.Duration {
	d, err := time.ParseDuration(duration)
	if err != nil {
		log.Fatalf("Invalid duration: %s", duration)
		os.Exit(1)
	}
	return d
}

func parseLogLevel(logLevel string) zerolog.Level {
	if logLevel == "" {
		return zerolog.InfoLevel
	}

	level, err := zerolog.ParseLevel(strings.ToLower(logLevel))
	if err != nil {
		log.Printf("Warning: invalid log level '%s', using 'info' as default", logLevel)
		return zerolog.InfoLevel
	}

	return level
}

func getEnvironment() string {
	env := strings.ToLower(strings.TrimSpace(os.Getenv("ENVIRONMENT")))
	if env == "" {
		return "dev"
	}
	return env
}

func validateEnvironment(env string) error {
	validEnvs := []string{"dev", "production", "prod"}

	for _, validEnv := range validEnvs {
		if env == validEnv {
			return nil
		}
	}

	return errors.New("environment must be one of: dev, production, prod")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseInt(value string) int {
	var result int
	_, err := fmt.Sscanf(value, "%d", &result)
	if err != nil {
		log.Printf("Warning: invalid integer value '%s', using 0", value)
		return 0
	}
	return result
}

func parseBool(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "true" || value == "1" || value == "yes"
}
