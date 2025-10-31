package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/skr1ms/dev2gis/config"
	"github.com/skr1ms/dev2gis/internal/auth"
	"github.com/skr1ms/dev2gis/internal/heightmap"
	"github.com/skr1ms/dev2gis/internal/storage"
	"github.com/skr1ms/dev2gis/internal/storage/minio"
	"github.com/skr1ms/dev2gis/pkg/jwt"
	"github.com/skr1ms/dev2gis/pkg/metrics"
	"github.com/skr1ms/dev2gis/pkg/middleware"
	"github.com/skr1ms/dev2gis/pkg/rabbitmq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           UAV Heights Map API
// @version         1.0
// @description     API for UAV data processing and 3D height map generation
// @contact.name    Dev2GIS Support
// @host            localhost:8080
// @BasePath        /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger := middleware.NewLogger(cfg)
	middleware.SetGlobalLogger(logger)

	ctx := context.Background()
	db, err := storage.NewDB(ctx, *cfg, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", err)
	}
	defer db.Close()

	minioClient, err := minio.NewMinio(cfg.Minio, logger)
	if err != nil {
		logger.Fatal("Failed to connect to MinIO", err)
	}

	rabbitmqClient, err := rabbitmq.NewClient(&cfg.RabbitMQ, logger)
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ", err)
	}
	defer rabbitmqClient.Close()

	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(logger.Middleware())
	router.Use(cors.Default())
	router.Use(metrics.MetricsMiddleware())

	healthHandler := func(c *gin.Context) {
		if err := db.HealthCheck(ctx); err != nil {
			logger.Error("Health check failed", err)
			c.JSON(500, gin.H{"status": "unhealthy", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "healthy", "environment": cfg.Server.Environment})
	}
	router.GET("/health", healthHandler)
	router.HEAD("/health", healthHandler)

	jwtService := jwt.NewJWTService(cfg.Auth.AccessTokenSecret, cfg.Auth.RefreshTokenSecret, cfg.Auth.JWTAccessTokenTTL, cfg.Auth.JWTRefreshTokenTTL)
	authService := auth.NewAuthService(db, jwtService)
	heightmapService := heightmap.NewService(db, minioClient, rabbitmqClient, cfg)

	heightmapHandler := heightmap.NewHandler(heightmapService)
	authHandler := auth.NewAuthHandler(authService, logger)

	// Register routes
	apiGroup := router.Group("/api")
	authHandler.RegisterRoutes(apiGroup)
	heightmapHandler.RegisterRoutes(apiGroup, middleware.NewJWTMiddleware(jwtService, logger))

	apiGroup.GET("/metrics", metrics.PrometheusHandler())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	port := ":" + cfg.Server.BackendPort
	logger.Info("Starting UAV Gateway server", map[string]interface{}{
		"port": port,
		"env":  cfg.Server.Environment,
		"mode": gin.Mode(),
	})

	if err := router.Run(port); err != nil {
		logger.Fatal("Failed to start server", err)
	}
}
