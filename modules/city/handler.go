package city

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

func NewHandler(cityService *Service) *Handler {
	return &Handler{service: cityService}
}

type CreateCityRequest struct {
	Name   string `json:"name" binding:"required,min=2,max=100"`
	Region string `json:"region" binding:"required,min=2,max=100"`
}

func (h *Handler) Create(cityRow *gin.Context) {
	var req CreateCityRequest
	if err := cityRow.ShouldBindJSON(&req); err != nil {
		cityRow.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.service.Create(req.Name, req.Region)
	if err != nil {
		if errors.Is(err, ErrNameTaken) {
			cityRow.JSON(http.StatusConflict, gin.H{"error": "name already created"})
			return
		}
		log.Printf("create failed: %v", err)
		cityRow.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	cityRow.JSON(http.StatusCreated, u)
}

type UpdateCityRequest struct {
	Name   string `json:"name" binding:"required,min=2,max=100"`
	Region string `json:"region" binding:"required,min=2,max=100"`
}

func (h *Handler) List(cityRow *gin.Context) {
	page, err := strconv.Atoi(cityRow.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(cityRow.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}
	search := cityRow.Query("search")

	cities, total, err := h.service.List(search, page, limit)
	if err != nil {
		log.Printf("list cities failed: %v", err)
		cityRow.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	cityRow.JSON(http.StatusOK, gin.H{
		"data": cities,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

func (h *Handler) GetByID(cityRow *gin.Context) {
	id, err := strconv.ParseInt(cityRow.Param("id"), 10, 64)
	if err != nil {
		cityRow.JSON(http.StatusBadRequest, gin.H{"error": "invalid city id"})
		return
	}

	u, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			cityRow.JSON(http.StatusNotFound, gin.H{"error": "city not found"})
			return
		}
		log.Printf("get city failed: %v", err)
		cityRow.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}
	cityRow.JSON(http.StatusOK, u)
}

func (h *Handler) Update(cityRow *gin.Context) {
	id, err := strconv.ParseInt(cityRow.Param("id"), 10, 64)
	if err != nil {
		cityRow.JSON(http.StatusBadRequest, gin.H{"error": "invalid city id"})
		return
	}

	var req UpdateCityRequest
	if err := cityRow.ShouldBindJSON(&req); err != nil {
		cityRow.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.service.Update(id, req.Name, req.Region)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			cityRow.JSON(http.StatusNotFound, gin.H{"error": "city not found"})
		case errors.Is(err, ErrNameTaken):
			cityRow.JSON(http.StatusConflict, gin.H{"error": "name already created"})
		default:
			log.Printf("update city failed: %v", err)
			cityRow.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		}
		return
	}
	cityRow.JSON(http.StatusOK, u)
}

func (h *Handler) Delete(cityRow *gin.Context) {
	id, err := strconv.ParseInt(cityRow.Param("id"), 10, 64)
	if err != nil {
		cityRow.JSON(http.StatusBadRequest, gin.H{"error": "invalid city id"})
		return
	}

	res, err := h.service.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, ErrCityHasMembers):
			// Fixed: changed 'c.JSON' to 'cityRow.JSON' to match your parameter
			cityRow.JSON(http.StatusConflict, gin.H{"error": "city still has active members"})

		case errors.Is(err, ErrNotFound):
			cityRow.JSON(http.StatusNotFound, gin.H{"error": "city not found"})

		default:
			log.Printf("delete city failed: %v", err)
			cityRow.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		}
		return
	}

	// Return 200 OK along with the JSON success data
	cityRow.JSON(http.StatusOK, res)
}
