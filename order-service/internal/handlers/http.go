package handlers

import (
	"errors"
	"log"
	"net/http"
	"order-service/internal/cache"
	"order-service/internal/data_base"
	"order-service/internal/models"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	db    *data_base.Postgres
	cache *cache.Cache
}

func New(db *data_base.Postgres, cache *cache.Cache) *Handler {
	h := &Handler{
		db:    db,
		cache: cache,
	}
	return h
}

type OrderResponse struct {
	Order      *models.Order `json:"order"`
	Source     string        `json:"source"` //"cache" or "db"
	DurationMs int64         `json:"duration"`
}

func (h *Handler) GetOrder(c *gin.Context) {
	startTime := time.Now()
	orderUID := c.Param("id")

	log.Printf("Processing request for order UID: '%s'", orderUID)

	// Валидация параметра
	if orderUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID not specified"})
		return
	}

	if len(orderUID) < 5 || len(orderUID) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID length"})
		return
	}

	// 1. Проверка кеша
	if cached := h.cache.Get(orderUID); cached != nil {
		log.Printf("Order %s found in cache", orderUID)
		response := models.NewOrderResponse(
			cached,
			"cache",
			time.Since(startTime),
		)
		c.JSON(http.StatusOK, response)
		return
	}

	log.Printf("Order %s not in cache, querying database...", orderUID)

	// 2. Запрос к БД
	dbStartTime := time.Now()
	order, err := h.db.GetOrderByUID(c.Request.Context(), orderUID)
	if err != nil {
		if errors.Is(err, data_base.ErrOrderNotFound) {
			log.Printf("Order %s not found in DB", orderUID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		} else {
			log.Printf("DB error for order %s: %v", orderUID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	log.Printf("Order %s found in DB, saving to cache...", orderUID)

	// 3. Сохранение в кеш

	if err := h.cache.Set(*order); err != nil {
		log.Printf("Failed to cache order %s: %v", orderUID, err)
	}

	response := models.NewOrderResponse(
		order,
		"database",
		time.Since(dbStartTime),
	)

	log.Printf("Successfully processed order %s in %v", orderUID, time.Since(startTime))
	c.JSON(http.StatusOK, response)
}
