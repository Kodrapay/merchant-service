package routes

import (
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/kodra-pay/merchant-service/internal/handlers"
	"github.com/kodra-pay/merchant-service/internal/repositories"
	"github.com/kodra-pay/merchant-service/internal/services"
)

func Register(app *fiber.App, serviceName string) {
	// Health check
	health := handlers.NewHealthHandler(serviceName)
	health.Register(app)

	// Get database URL from environment
	dbURL := os.Getenv("POSTGRES_URL")
	if dbURL == "" {
		dbURL = "postgres://kodrapay:kodrapay_password@localhost:5432/kodrapay?sslmode=disable"
	} else {
		// Add sslmode=disable if not already present
		if !strings.Contains(dbURL, "sslmode=") {
			dbURL = dbURL + "?sslmode=disable"
		}
	}

	// Initialize repository
	repo, err := repositories.NewMerchantRepository(dbURL)
	if err != nil {
		log.Printf("Warning: Failed to connect to database: %v. Using stub implementation.", err)
		// Continue with stub implementation
		return
	}

	// Initialize service
	merchantService := services.NewMerchantService(repo)

	// Initialize handlers
	merchantHandler := handlers.NewMerchantHandler(merchantService)
	kycHandler := handlers.NewKYCHandler(merchantService)

	// Register routes
	merchantHandler.Register(app)
	kycHandler.Register(app)
}
