package member

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

type CreateMemberRequest struct {
	FullName string `json:"full_name" binding:"required,min=2,max=150"`
	Phone    string `json:"phone" binding:"required,min=7,max=20"`
	CityID   int64  `json:"city_id" binding:"required,gt=0"`
}

type UpdateMemberRequest struct {
	FullName string `json:"full_name" binding:"required,min=2,max=150"`
	Phone    string `json:"phone" binding:"required,min=7,max=20"`
	CityID   int64  `json:"city_id" binding:"required,gt=0"`
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m, err := h.service.Create(req.FullName, req.Phone, req.CityID)
	if err != nil {
		switch {
		case errors.Is(err, ErrPhoneTaken):
			c.JSON(http.StatusConflict, gin.H{"error": "phone already registered"})
		case errors.Is(err, ErrCityInvalid):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "city does not exist"})
		default:
			log.Printf("create member failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		}
		return
	}
	c.JSON(http.StatusCreated, m)
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

	cityID, err := strconv.ParseInt(c.DefaultQuery("city_id", "0"), 10, 64)
	if err != nil || cityID < 0 {
		cityID = 0
	}

	members, total, err := h.service.List(search, cityID, page, limit)
	if err != nil {
		log.Printf("list members failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": members,
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member id"})
		return
	}

	m, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
			return
		}
		log.Printf("get member failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member id"})
		return
	}

	var req UpdateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m, err := h.service.Update(id, req.FullName, req.Phone, req.CityID)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
		case errors.Is(err, ErrPhoneTaken):
			c.JSON(http.StatusConflict, gin.H{"error": "phone already registered"})
		case errors.Is(err, ErrCityInvalid):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "city does not exist"})
		default:
			log.Printf("update member failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		}
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member id"})
		return
	}

	if err := h.service.Delete(id); err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
			return
		}
		log.Printf("delete member failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}
	c.Status(http.StatusNoContent)
}
