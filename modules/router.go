package modules

import (
	"database/sql"
	"net/http"

	"financial-system/config"
	"financial-system/middleware"
	"financial-system/modules/charge"
	"financial-system/modules/city"
	"financial-system/modules/dashboard"
	"financial-system/modules/feeTypes"
	"financial-system/modules/member"
	"financial-system/modules/payment"
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

	feeTypeRepo := feeTypes.NewRepository(dbConn)
	feeTypeService := feeTypes.NewService(feeTypeRepo)
	feeTypeHandler := feeTypes.NewHandler(feeTypeService)

	chargeRepo := charge.NewRepository(dbConn)
	chargeService := charge.NewService(chargeRepo, memberService, feeTypeService)
	chargeHandler := charge.NewHandler(chargeService)

	paymentRepo := payment.NewRepository(dbConn)
	paymentService := payment.NewService(paymentRepo, chargeService)
	paymentHandler := payment.NewHandler(paymentService)

	dashboardRepo := dashboard.NewRepository(dbConn)
	dashboardService := dashboard.NewService(dashboardRepo)
	dashboardHandler := dashboard.NewHandler(dashboardService)

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

			feeTypes := protected.Group("/feeTypes")
			{
				feeTypes.POST("", feeTypeHandler.Create)
				feeTypes.GET("", feeTypeHandler.List)
				feeTypes.GET("/:id", feeTypeHandler.GetByID)
				feeTypes.PUT("/:id", feeTypeHandler.Update)
				feeTypes.DELETE("/:id", feeTypeHandler.Delete)
			}

			charges := protected.Group("/monthly-charges")
			{
				charges.POST("", chargeHandler.Create)
				charges.GET("", chargeHandler.List)
				charges.GET("/:id", chargeHandler.GetByID)
				charges.PUT("/:id", chargeHandler.Update)
				charges.DELETE("/:id", chargeHandler.Delete)
			}

			payments := protected.Group("/payments")
			{
				payments.POST("", paymentHandler.Record)
				payments.GET("", paymentHandler.List)
				payments.GET("/:id", paymentHandler.GetByID)
				payments.GET("/summary", paymentHandler.Summary)
			}

			dash := protected.Group("/dashboard")
			{
				dash.GET("/summary", dashboardHandler.Summary)
				dash.GET("/monthly-revenue", dashboardHandler.MonthlyRevenue)
				dash.GET("/revenue-by-city", dashboardHandler.RevenueByCity)
				dash.GET("/revenue-by-fee-type", dashboardHandler.RevenueByFeeType)
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
