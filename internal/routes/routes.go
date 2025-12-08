package routes

import (
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/kodra-pay/merchant-service/internal/clients"
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

	// Wallet-ledger service base URL
	walletLedgerURL := os.Getenv("WALLET_LEDGER_SERVICE_URL")
	if walletLedgerURL == "" {
		walletLedgerURL = "http://wallet-ledger-service:7007/api/v1"
	}
	walletLedgerClient := clients.NewHTTPWalletLedgerClient(walletLedgerURL)

	// Get database connection from merchant repository for other repos
	db := merchantRepo.GetDB()
	paymentOptionsRepo := repositories.NewPaymentOptionsRepository(db)
	settlementConfigRepo := repositories.NewSettlementConfigRepository(db)
	paymentLinkRepo := repositories.NewPaymentLinkRepository(db)
	apiKeyRepo := repositories.NewAPIKeyRepository(db)
	kycSubmissionRepo := repositories.NewKYCSubmissionRepository(db)
	balanceRepo := repositories.NewBalanceRepository(db)

	// Initialize services
	merchantService := services.NewMerchantService(merchantRepo, apiKeyRepo, settlementConfigRepo, walletLedgerClient)
	kycService := services.NewKYCService(merchantRepo, kycSubmissionRepo)
	paymentOptionsService := services.NewPaymentOptionsService(paymentOptionsRepo)
	settlementConfigService := services.NewSettlementConfigService(settlementConfigRepo)
	paymentLinkService := services.NewPaymentLinkService(paymentLinkRepo)
	balanceService := services.NewBalanceService(balanceRepo)

	// Initialize handlers
	merchantHandler := handlers.NewMerchantHandler(merchantService)
	kycHandler := handlers.NewKYCHandler(merchantService, kycService)
	paymentOptionsHandler := handlers.NewPaymentOptionsHandler(paymentOptionsService, settlementConfigService)
	paymentLinkHandler := handlers.NewPaymentLinkHandler(paymentLinkService)
	balanceHandler := handlers.NewBalanceHandler(balanceService)

	// Register routes
	merchantHandler.Register(app)
	kycHandler.Register(app)
	paymentOptionsHandler.Register(app)
	paymentLinkHandler.Register(app)
	balanceHandler.Register(app)
}
