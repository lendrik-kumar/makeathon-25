package controllers

import (
	"net/http"

	"backend/models"

	"github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "errors"
	"encoding/json"
	"fmt"
)

func InitProductController(database *gorm.DB) {
	db = database
}

type ProductInput struct {
	SerialNumber string `json:"serial_number" binding:"required"`
	Manufacturer string `json:"manufacturer" binding:"required"`
	Model        string `json:"model" binding:"required"`
}

func RegisterProduct(c *gin.Context) {
	var input ProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.String(http.StatusBadRequest, "failed to bind JSON: %v", err)
		return
	}

	product := models.Product{
		SerialNumber: input.SerialNumber,
		Manufacturer: input.Manufacturer,
		ProductModel: input.Model,
	}

	// create the product in db
	if err := db.Create(&product).Error; err != nil {
		c.String(http.StatusInternalServerError, "failed to create product: %v", err)
		return
	}

	userID := c.MustGet("user_id")

	// create the event in db
	event := models.Event{
		ProductID: product.ID,
		EventType: "registration",
		EventData: `{"details": "Product registered"}`,
		CreatedBy: userID.(uint),
	}

	if err := createEvent(&event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log registration event"})
		return
	}

	c.JSON(http.StatusOK, product)
}

func GetProduct(c *gin.Context) {
	id := c.Param("id")
	var product models.Product
	if err := db.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	var events []models.Event
	db.Where("product_id = ?", id).Order("created_at asc").Find(&events)

	c.JSON(http.StatusOK, gin.H{
		"product": product,
		"history": events,
	})
}

type TransferInput struct {
    NewOwnerUsername string `json:"new_owner_username" binding:"required"`
}

func InitiateTransfer(c *gin.Context) {
    productID := c.Param("id")
    userID, _ := c.Get("user_id")

    var product models.Product
    if err := db.First(&product, productID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
        return
    }

    var lastEvent models.Event
    if err := db.Where("product_id = ? AND event_type = ?", productID, "ownership_transfer").Order("created_at desc").First(&lastEvent).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    var currentOwnerID uint
    if lastEvent.ID != 0 {
        var eventData map[string]interface{}
        json.Unmarshal([]byte(lastEvent.EventData), &eventData)
        currentOwnerID = uint(eventData["new_owner_id"].(float64))
    } else {
        // Assume initial owner is the one who registered it
        var regEvent models.Event
        if err := db.Where("product_id = ? AND event_type = ?", productID, "registration").First(&regEvent).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "No registration event found"})
            return
        }
        currentOwnerID = regEvent.CreatedBy
    }

    if currentOwnerID != userID.(uint) {
        c.JSON(http.StatusForbidden, gin.H{"error": "Only the current owner can initiate a transfer"})
        return
    }

    var input TransferInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var newOwner models.User
    if err := db.Where("username = ?", input.NewOwnerUsername).First(&newOwner).Error; err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "New owner not found"})
        return
    }

    pendingTransfer := models.PendingTransfer{
        ProductID:  product.ID,
        NewOwnerID: newOwner.ID,
    }

    if err := db.Create(&pendingTransfer).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate transfer"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Transfer initiated"})
}

func ConfirmTransfer(c *gin.Context) {
    productID := c.Param("id")
    userID, _ := c.Get("user_id")

    var pendingTransfer models.PendingTransfer
    if err := db.Where("product_id = ? AND new_owner_id = ?", productID, userID).First(&pendingTransfer).Error; err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "No pending transfer found"})
        return
    }

    event := models.Event{
        ProductID:  pendingTransfer.ProductID,
        EventType:  "ownership_transfer",
        EventData:  fmt.Sprintf(`{"new_owner_id": %d}`, pendingTransfer.NewOwnerID),
        CreatedBy:  userID.(uint),
    }

    if err := createEvent(&event); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log transfer event"})
        return
    }

    if err := db.Delete(&pendingTransfer).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete pending transfer"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Transfer confirmed"})
}
