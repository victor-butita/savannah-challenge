package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/victor-butita/savannah-challenge/internal/auth"
	"github.com/victor-butita/savannah-challenge/internal/database"
	"github.com/victor-butita/savannah-challenge/internal/models"
	"gorm.io/gorm"
)

type MockSMSService struct {
	mock.Mock
	wg *sync.WaitGroup
}

func (m *MockSMSService) Send(recipient, message string) {
	m.Called(recipient, message)
	if m.wg != nil {
		m.wg.Done()
	}
}

type MockOIDCVerifier struct {
	mock.Mock
}

func (m *MockOIDCVerifier) Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	args := m.Called(ctx, rawIDToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*oidc.IDToken), args.Error(1)
}

func setupTestRouter() (*gin.Engine, *gorm.DB, *MockSMSService, *MockOIDCVerifier) {
	gin.SetMode(gin.TestMode)
	db, err := database.Connect("file::memory:")
	if err != nil {
		panic("Failed to connect to in-memory database")
	}

	mockSMS := &MockSMSService{}
	mockVerifier := new(MockOIDCVerifier)

	handler := NewHandler(db, mockSMS)
	authMiddleware := auth.NewAuthMiddleware(mockVerifier)

	router := gin.Default()
	apiV1 := router.Group("/api/v1")
	{
		apiV1.POST("/customers", handler.CreateCustomer)

		authorized := apiV1.Group("/")
		authorized.Use(authMiddleware.ValidateToken())
		{
			authorized.POST("/orders", handler.CreateOrder)
		}
	}
	return router, db, mockSMS, mockVerifier
}

func TestCreateCustomer(t *testing.T) {
	router, _, _, _ := setupTestRouter()

	customerReq := CreateCustomerRequest{
		Name:        "Test User",
		Code:        "TU001",
		PhoneNumber: "+254700000000",
	}
	body, _ := json.Marshal(customerReq)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/customers", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var customer models.Customer
	json.Unmarshal(w.Body.Bytes(), &customer)
	assert.Equal(t, customerReq.Name, customer.Name)
	assert.Equal(t, customerReq.Code, customer.Code)
}

func TestCreateOrder(t *testing.T) {
	router, db, mockSMS, mockVerifier := setupTestRouter()

	customer := models.Customer{Name: "Test User", Code: "TU001", PhoneNumber: "+254700000000"}
	db.Create(&customer)

	orderReq := CreateOrderRequest{
		Item:       "Laptop",
		Amount:     1500.00,
		CustomerID: customer.ID,
	}
	body, _ := json.Marshal(orderReq)

	mockVerifier.On("Verify", mock.Anything, "valid-token").Return(&oidc.IDToken{}, nil)

	var wg sync.WaitGroup
	wg.Add(1)
	mockSMS.wg = &wg
	mockSMS.On("Send", "+254700000000", "Dear Test User, your order for Laptop has been received.").Once()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	router.ServeHTTP(w, req)

	wg.Wait() // wait until Send is called

	assert.Equal(t, http.StatusCreated, w.Code)
	mockSMS.AssertExpectations(t)
	mockVerifier.AssertExpectations(t)
}
