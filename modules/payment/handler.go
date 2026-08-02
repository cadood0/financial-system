package payment

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const periodLayout = "2006-01"

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type RecordPaymentRequest struct {
	ChargeID    int64  `json:"charge_id" binding:"required,gt=0"`
	Period      string `json:"period" binding:"required"`
	AmountCents int64  `json:"amount_cents" binding:"required,gt=0"`
	Note        string `json:"note" binding:"max=250"`
}

func (h *Handler) Record(c *gin.Context) {
	var req RecordPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	period, err := time.Parse(periodLayout, req.Period)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "period must be YYYY-MM"})
		return
	}

	recordedBy := c.GetInt64("user_id")

	p, err := h.service.Record(req.ChargeID, req.AmountCents, recordedBy, period, req.Note)
	if err != nil {
		switch {
		case errors.Is(err, ErrChargeInvalid):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "monthly charge does not exist"})
		case errors.Is(err, ErrPeriodOutOfRange):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "period is outside the charge's active range"})
		case errors.Is(err, ErrOverpayment):
			c.JSON(http.StatusConflict, gin.H{"error": "payment exceeds remaining balance for this period"})
		default:
			log.Printf("record payment failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		}
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *Handler) List(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}
	memberID, err := strconv.ParseInt(c.DefaultQuery("member_id", "0"), 10, 64)
	if err != nil || memberID < 0 {
		memberID = 0
	}
	chargeID, err := strconv.ParseInt(c.DefaultQuery("charge_id", "0"), 10, 64)
	if err != nil || chargeID < 0 {
		chargeID = 0
	}

	payments, total, err := h.service.List(memberID, chargeID, page, limit)
	if err != nil {
		log.Printf("list payments failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": payments,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment id"})
		return
	}

	p, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}
		log.Printf("get payment failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handler) Summary(c *gin.Context) {
	chargeID, err := strconv.ParseInt(c.Query("charge_id"), 10, 64)
	if err != nil || chargeID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "charge_id is required"})
		return
	}

	period, err := time.Parse(periodLayout, c.Query("period"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "period must be YYYY-MM"})
		return
	}

	summary, err := h.service.Summary(chargeID, period)
	if err != nil {
		if errors.Is(err, ErrChargeInvalid) {
			c.JSON(http.StatusNotFound, gin.H{"error": "monthly charge does not exist"})
			return
		}
		log.Printf("payment summary failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}
	c.JSON(http.StatusOK, summary)
}
