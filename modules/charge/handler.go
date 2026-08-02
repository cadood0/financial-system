package charge

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const dateLayout = "2006-01-02"

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type CreateChargeRequest struct {
	MemberID    int64   `json:"member_id" binding:"required,gt=0"`
	FeeTypeID   int64   `json:"fee_type_id" binding:"required,gt=0"`
	AmountCents int64   `json:"amount_cents" binding:"required,gt=0"`
	StartDate   string  `json:"start_date" binding:"required"`
	EndDate     *string `json:"end_date"`
}

type UpdateChargeRequest struct {
	AmountCents int64   `json:"amount_cents" binding:"required,gt=0"`
	EndDate     *string `json:"end_date"`
}

func parseDate(s string) (time.Time, error) {
	return time.Parse(dateLayout, s)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateChargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startDate, err := parseDate(req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date must be YYYY-MM-DD"})
		return
	}

	var endDate *time.Time
	if req.EndDate != nil {
		parsed, err := parseDate(*req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "end_date must be YYYY-MM-DD"})
			return
		}
		endDate = &parsed
	}

	mc, err := h.service.Create(req.MemberID, req.FeeTypeID, req.AmountCents, startDate, endDate)
	if err != nil {
		switch {
		case errors.Is(err, ErrMemberInvalid):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "member does not exist"})
		case errors.Is(err, ErrFeeTypeInvalid):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "fee type does not exist or is inactive"})
		case errors.Is(err, ErrDuplicateCharge):
			c.JSON(http.StatusConflict, gin.H{"error": "member already has an active charge for this fee type"})
		case errors.Is(err, ErrBadDates):
			c.JSON(http.StatusBadRequest, gin.H{"error": "end date must be after start date"})
		default:
			log.Printf("create charge failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		}
		return
	}
	c.JSON(http.StatusCreated, mc)
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
	feeTypeID, err := strconv.ParseInt(c.DefaultQuery("fee_type_id", "0"), 10, 64)
	if err != nil || feeTypeID < 0 {
		feeTypeID = 0
	}

	charges, total, err := h.service.List(memberID, feeTypeID, page, limit)
	if err != nil {
		log.Printf("list charges failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": charges,
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid charge id"})
		return
	}

	mc, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "monthly charge not found"})
			return
		}
		log.Printf("get charge failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}
	c.JSON(http.StatusOK, mc)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid charge id"})
		return
	}

	var req UpdateChargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var endDate *time.Time
	if req.EndDate != nil {
		parsed, err := parseDate(*req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "end_date must be YYYY-MM-DD"})
			return
		}
		endDate = &parsed
	}

	mc, err := h.service.Update(id, req.AmountCents, endDate)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "monthly charge not found"})
		case errors.Is(err, ErrBadDates):
			c.JSON(http.StatusBadRequest, gin.H{"error": "end date must be after start date"})
		default:
			log.Printf("update charge failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		}
		return
	}
	c.JSON(http.StatusOK, mc)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid charge id"})
		return
	}

	if err := h.service.Delete(id); err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "monthly charge not found"})
			return
		}
		log.Printf("delete charge failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}
	c.Status(http.StatusNoContent)
}
