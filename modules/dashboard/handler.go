package dashboard

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Summary(c *gin.Context) {
	s, err := h.service.Summary()
	if err != nil {
		log.Printf("dashboard summary failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *Handler) MonthlyRevenue(c *gin.Context) {
	months, err := strconv.Atoi(c.DefaultQuery("months", "12"))
	if err != nil || months < 1 || months > 60 {
		months = 12
	}

	result, err := h.service.MonthlyRevenue(months)
	if err != nil {
		log.Printf("monthly revenue failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *Handler) RevenueByCity(c *gin.Context) {
	result, err := h.service.RevenueByCity()
	if err != nil {
		log.Printf("revenue by city failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *Handler) RevenueByFeeType(c *gin.Context) {
	result, err := h.service.RevenueByFeeType()
	if err != nil {
		log.Printf("revenue by fee type failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}
