package modules

import (
	"database/sql"
	"net/http"

	"financial-system/config"
	"financial-system/middleware"
	"financial-system/modules/city"
	"financial-system/modules/member"
	"financial-system/modules/user"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, dbConn *sql.DB, cfg *config.Config) {
	router.GET("/health", healthCheck(dbConn))

	userRepo := user.NewRepository(dbConn)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService, cfg)

	memberRepo := member.NewRepository(dbConn)

	cityRepo := city.NewRepository(dbConn)
	cityService := city.NewService(cityRepo, memberRepo)
	cityHandler := city.NewHandler(cityService)

	memberService := member.NewService(memberRepo, cityService)
	memberHandler := member.NewHandler(memberService)

	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}

		protected := v1.Group("")
		protected.Use(middleware.Auth(cfg.JWTSecret))
		{
			protected.GET("/me", userHandler.Me)

			users := protected.Group("/users")
			{
				users.GET("", userHandler.List)
				users.GET("/:id", userHandler.GetByID)
				users.PUT("/:id", userHandler.Update)
				users.DELETE("/:id", userHandler.Delete)
			}

			cities := protected.Group("/cities")
			{
				cities.POST("", cityHandler.Create)
				cities.GET("", cityHandler.List)
				cities.GET("/:id", cityHandler.GetByID)
				cities.PUT("/:id", cityHandler.Update)
				cities.DELETE("/:id", cityHandler.Delete)
			}

			members := protected.Group("/members")
			{
				members.POST("", memberHandler.Create)
				members.GET("", memberHandler.List)
				members.GET("/:id", memberHandler.GetByID)
				members.PUT("/:id", memberHandler.Update)
				members.DELETE("/:id", memberHandler.Delete)
			}
		}
	}
}

func healthCheck(dbConn *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := dbConn.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "database": "down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "database": "up"})
	}
}
