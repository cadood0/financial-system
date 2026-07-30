package feetype

import (
	"errors"
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

type CreateFeeTypeRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=150"`
	Description string `json:"description" binding:"max=250"`
	IsActive    *bool  `json:"is_active"`
}

type UpdateFeeTypeRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=150"`
	Description string `json:"description" binding:"max=250"`
	IsActive    *bool  `json:"is_active"`
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateFeeTypeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	active := true

	if req.IsActive != nil {
		active = *req.IsActive
	}

	feeType, err := h.service.Create(
		req.Name,
		req.Description,
		active,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrNameTaken):
			c.JSON(http.StatusConflict, gin.H{"error": "fee type already exists"})
		default:
			log.Printf("create fee type failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		}
		return
	}

	c.JSON(http.StatusCreated, feeType)
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

	search := c.Query("search")

	feeTypes, total, err := h.service.List(search, page, limit)
	if err != nil {
		log.Printf("list fee types failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": feeTypes,
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid fee type id"})
		return
	}

	feeType, err := h.service.GetByID(id)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "fee type not found"})
		default:
			log.Printf("get fee type failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		}
		return
	}

	c.JSON(http.StatusOK, feeType)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid fee type id"})
		return
	}

	var req UpdateFeeTypeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current record
	current, err := h.service.GetByID(id)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "fee type not found"})
		default:
			log.Printf("get fee type failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		}
		return
	}

	// Keep existing value unless client sent a new one
	active := current.IsActive
	if req.IsActive != nil {
		active = *req.IsActive
	}

	feeType, err := h.service.Update(
		id,
		req.Name,
		req.Description,
		active,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "fee type not found"})
		case errors.Is(err, ErrNameTaken):
			c.JSON(http.StatusConflict, gin.H{"error": "fee type already exists"})
		default:
			log.Printf("update fee type failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		}
		return
	}

	c.JSON(http.StatusOK, feeType)
}
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid fee type id"})
		return
	}

	if err := h.service.Delete(id); err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "fee type not found"})
		default:
			log.Printf("delete fee type failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
