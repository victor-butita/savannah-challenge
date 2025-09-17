package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/victor-butita/savannah-challenge/internal/models"
	"github.com/victor-butita/savannah-challenge/internal/services"
	"gorm.io/gorm"
)

type Handler struct {
	DB         *gorm.DB
	SMSService services.SMSServiceInterface
}

func NewHandler(db *gorm.DB, smsService services.SMSServiceInterface) *Handler {
	return &Handler{DB: db, SMSService: smsService}
}

type CreateCustomerRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required"`
}

func (h *Handler) CreateCustomer(c *gin.Context) {
	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer := models.Customer{
		Name:        req.Name,
		Code:        req.Code,
		PhoneNumber: req.PhoneNumber,
	}

	if result := h.DB.Create(&customer); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, customer)
}

type CreateOrderRequest struct {
	Item       string  `json:"item" binding:"required"`
	Amount     float64 `json:"amount" binding:"required"`
	CustomerID uint    `json:"customer_id" binding:"required"`
}

func (h *Handler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var customer models.Customer
	if err := h.DB.First(&customer, req.CustomerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
		return
	}

	order := models.Order{
		Item:       req.Item,
		Amount:     req.Amount,
		Time:       time.Now().UTC(),
		CustomerID: req.CustomerID,
	}

	if result := h.DB.Create(&order); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	go func() {
		message := fmt.Sprintf("Dear %s, your order for %s has been received.", customer.Name, order.Item)
		h.SMSService.Send(customer.PhoneNumber, message)
	}()

	c.JSON(http.StatusCreated, order)
}
