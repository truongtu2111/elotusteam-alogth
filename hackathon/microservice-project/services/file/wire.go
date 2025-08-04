package main

import (
	"fmt"
	"log"
	"time"

	"github.com/elotusteam/microservice-project/services/file/clients"
	"github.com/elotusteam/microservice-project/services/file/infrastructure"
	"github.com/elotusteam/microservice-project/services/file/usecases"
	"github.com/elotusteam/microservice-project/shared/config"
)

// ServiceContainer holds all the services
type ServiceContainer struct {
	FileService            usecases.FileService
	ImageProcessingService usecases.ImageProcessingService
}

// NewServiceContainer creates and wires all dependencies
func NewServiceContainer(cfg *config.Config) (*ServiceContainer, error) {
	// Create default config if none provided
	if cfg == nil {
		cfg = &config.Config{
			Storage: config.StorageConfig{
				LocalPath: "/tmp/file-service",
				CDN: config.CDNConfig{
					Enabled: false,
					BaseURL: "http://localhost:8080",
				},
			},
			Database: config.DatabaseConfig{
				Driver:             "postgres",
				Host:               "localhost",
				Port:               5432,
				Database:           "file_db",
				Username:           "postgres",
				Password:           "password",
				SSLMode:            "disable",
				MaxOpenConnections: 25,
				MaxIdleConnections: 5,
				ConnectionTimeout:  30 * time.Second,
				QueryTimeout:       30 * time.Second,
				ConnectionLifetime: 5 * time.Minute,
			},
		}
	}

	// Create database connection
	db, err := infrastructure.NewPostgreSQLConnection(&cfg.Database)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Create repository manager
	repoManager := infrastructure.NewPostgreSQLRepositoryManager(db)

	// Create storage service
	storageService, err := infrastructure.NewMultiCDNStorageService(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage service: %w", err)
	}

	// Create image processing service
	imageProcessingService := infrastructure.NewImageProcessingService(
		repoManager,
		storageService,
		cfg,
	)

	// Create real HTTP client services
	permissionService := clients.NewPermissionClient(cfg.Services.Auth.BaseURL)
	notificationService := clients.NewNotificationClient(cfg.Services.Notification.BaseURL)
	activityService := clients.NewActivityClient(cfg.Services.Analytics.BaseURL)

	// Create file service
	fileService := usecases.NewFileService(
		repoManager,
		storageService,
		permissionService,
		notificationService,
		activityService,
		imageProcessingService,
		cfg,
	)

	return &ServiceContainer{
		FileService:            fileService,
		ImageProcessingService: imageProcessingService,
	}, nil
}
