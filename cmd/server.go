package cmd

import (
	"log"

	"financial-system/config"
	"financial-system/db"
	"financial-system/modules"

	"github.com/gin-gonic/gin"
)

func Start(cfg *config.Config) {
	database, err := db.Connect(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer database.Close()
	log.Println("database connected")

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	modules.SetupRoutes(router, database, cfg)

	log.Printf("starting %s on port %s", cfg.AppName, cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
