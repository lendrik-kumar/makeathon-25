package controllers

import (
	"backend/models"
	"backend/utils"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitEventController(database *gorm.DB) {
	db = database
}

type EventInput struct {
	EventType string `json:"event_type" binding:"required"`
	EventData string `json:"event_data" binding:"required"`
}

func createEvent(c *gin.Context) {
	productID := c.Param("id")
	role, _ := c.Get("role")

	var input EventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.EventType == "repair" && role != "repair_shop" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only repair shops can log repairs"})
		return
	}

	var lastEvent models.Event
	err := db.Where("product_id = ?", productID).Order("created_at desc").First(&lastEvent).Error
	previousHash := ""
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		previousHash = lastEvent.EventHash
	}

	productIDUint, err := strconv.ParseUint(productID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	userID, _ := c.Get("user_id")
	event := models.Event{
		ProductID:         uint(productIDUint),
		EventType:         input.EventType,
		EventData:         input.EventData,
		PreviousEventHash: previousHash,
		CreatedBy:         userID.(uint),
	}

	if err := db.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}

	hashData := utils.EventHashData{
		ProductID:         event.ProductID,
		EventType:         event.EventType,
		EventData:         event.EventData,
		CreatedAt:         event.CreatedAt,
		CreatedBy:         event.CreatedBy,
		PreviousEventHash: event.PreviousEventHash,
	}

	eventHash, err := utils.ComputeEventHash(hashData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to compute hash"})
		return
	}

	event.EventHash = eventHash
	if err := db.Save(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event hash"})
		return
	}

	c.JSON(http.StatusOK, event)
}

func VerifyProductHistory(c *gin.Context) {
	productID := c.Param("id")
	var events []models.Event
	if err := db.Where("product_id = ?", productID).Order("created_at asc").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for i, event := range events {
		if i == 0 && event.PreviousEventHash != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid previous hash for first event"})
			return
		} else if i > 0 && event.PreviousEventHash != events[i-1].EventHash {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Hash chain broken at event " + string(event.ID)})
			return
		}

		hashData := utils.EventHashData{
			ProductID:         event.ProductID,
			EventType:         event.EventType,
			EventData:         event.EventData,
			CreatedAt:         event.CreatedAt,
			CreatedBy:         event.CreatedBy,
			PreviousEventHash: event.PreviousEventHash,
		}

		expectedHash, err := utils.ComputeEventHash(hashData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to compute hash"})
			return
		}

		if event.EventHash != expectedHash {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hash for event " + string(event.ID)})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "History is valid"})
}
