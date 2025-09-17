package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/victor-butita/savannah-challenge/internal/api"
	"github.com/victor-butita/savannah-challenge/internal/auth"
	"github.com/victor-butita/savannah-challenge/internal/config"
	"github.com/victor-butita/savannah-challenge/internal/database"
	"github.com/victor-butita/savannah-challenge/internal/services"
)

func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}

	smsService := services.NewSMSService(cfg.ATUsername, cfg.ATAPIKey, cfg.ATEnv)
	oidcVerifier, err := auth.NewOIDCVerifier(cfg.OIDCProviderURL, cfg.OIDCClientID)
	if err != nil {
		log.Fatalf("could not create OIDC verifier: %v", err)
	}

	handler := api.NewHandler(db, smsService)
	authMiddleware := auth.NewAuthMiddleware(oidcVerifier)

	router := gin.Default()

	apiV1 := router.Group("/api/v1")
	{
		apiV1.POST("/customers", handler.CreateCustomer)

		authorized := apiV1.Group("/")
		authorized.Use(authMiddleware.ValidateToken())
		{
			authorized.POST("/orders", handler.CreateOrder)
		}
	}

	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
