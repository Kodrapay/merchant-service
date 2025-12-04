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

	// Initialize repositories
	merchantRepo, err := repositories.NewMerchantRepository(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Get database connection from merchant repository for other repos
	db := merchantRepo.GetDB()
	paymentOptionsRepo := repositories.NewPaymentOptionsRepository(db)
	settlementConfigRepo := repositories.NewSettlementConfigRepository(db)

	// Initialize services
	merchantService := services.NewMerchantService(merchantRepo)
	paymentOptionsService := services.NewPaymentOptionsService(paymentOptionsRepo)
	settlementConfigService := services.NewSettlementConfigService(settlementConfigRepo)

	// Initialize handlers
	merchantHandler := handlers.NewMerchantHandler(merchantService)
	kycHandler := handlers.NewKYCHandler(merchantService)
	paymentOptionsHandler := handlers.NewPaymentOptionsHandler(paymentOptionsService, settlementConfigService)

	// Register routes
	merchantHandler.Register(app)
	kycHandler.Register(app)
	paymentOptionsHandler.Register(app)
}
