package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/25x8/ya-prakt-6-sprint/internal/gophermart/middleware"
	"github.com/25x8/ya-prakt-6-sprint/internal/gophermart/models"
	"github.com/25x8/ya-prakt-6-sprint/internal/gophermart/repository"
	"github.com/25x8/ya-prakt-6-sprint/internal/gophermart/service"
	"github.com/25x8/ya-prakt-6-sprint/internal/gophermart/utils"
	"golang.org/x/crypto/bcrypt"
)

// Handler handles all HTTP requests
type Handler struct {
	Repo       repository.Repository
	AccrualSvc *service.AccrualService
	JWTSecret  string
}

// NewHandler creates a new handler
func NewHandler(repo repository.Repository, accrualSvc *service.AccrualService, jwtSecret string) *Handler {
	return &Handler{
		Repo:       repo,
		AccrualSvc: accrualSvc,
		JWTSecret:  jwtSecret,
	}
}

// RegisterUser handles user registration
func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	// Parse request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	ctx := r.Context()
	existingUser, err := h.Repo.GetUserByLogin(ctx, req.Login)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if existingUser != nil {
		http.Error(w, "Login already taken", http.StatusConflict)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Create user
	userID, err := h.Repo.CreateUser(ctx, req.Login, string(hashedPassword))
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Generate token
	token, err := middleware.GenerateToken(userID, h.JWTSecret)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Set cookie and header
	middleware.SetAuthCookie(w, token)
	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

// LoginUser handles user login
func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	// Parse request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	// Get user
	ctx := r.Context()
	user, err := h.Repo.GetUserByLogin(ctx, req.Login)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate token
	token, err := middleware.GenerateToken(user.ID, h.JWTSecret)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Set cookie and header
	middleware.SetAuthCookie(w, token)
	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

// UploadOrder handles order upload
func (h *Handler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Read order number
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	orderNumber := string(body)
	if orderNumber == "" {
		http.Error(w, "Empty order number", http.StatusBadRequest)
		return
	}

	// Validate order number with Luhn algorithm
	if !utils.IsNumeric(orderNumber) || !utils.ValidateLuhn(orderNumber) {
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	ctx := r.Context()

	// Check if order already exists
	existingOrder, err := h.Repo.GetOrderByNumber(ctx, orderNumber)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// If order exists and belongs to this user, return 200
	if existingOrder != nil && existingOrder.UserID == userID {
		w.WriteHeader(http.StatusOK)
		return
	}

	// If order exists but belongs to another user, return 409
	if existingOrder != nil {
		http.Error(w, "Order already uploaded by another user", http.StatusConflict)
		return
	}

	// Create order
	err = h.Repo.CreateOrder(ctx, userID, orderNumber)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Start processing the order (in real implementation, this should be done in background)
	go h.processOrder(orderNumber)

	w.WriteHeader(http.StatusAccepted)
}

// GetOrders returns the list of user's orders
func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get orders
	ctx := r.Context()
	orders, err := h.Repo.GetUserOrders(ctx, userID)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// If no orders, return 204
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Prepare response
	type orderResponse struct {
		Number     string    `json:"number"`
		Status     string    `json:"status"`
		Accrual    float64   `json:"accrual,omitempty"`
		UploadedAt time.Time `json:"uploaded_at"`
	}

	response := make([]orderResponse, 0, len(orders))
	for _, order := range orders {
		orderResp := orderResponse{
			Number:     order.Number,
			Status:     order.Status,
			UploadedAt: order.UploadedAt,
		}

		// Only include accrual if status is PROCESSED
		if order.Status == models.StatusProcessed {
			orderResp.Accrual = order.Accrual
		}

		response = append(response, orderResp)
	}

	// Return orders
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetBalance returns user's balance
func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get balance
	ctx := r.Context()
	balance, err := h.Repo.GetUserBalance(ctx, userID)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Return balance
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}

// WithdrawBalance handles balance withdrawal
func (h *Handler) WithdrawBalance(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Order string  `json:"order"`
		Sum   float64 `json:"sum"`
	}

	// Parse request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Validate order number with Luhn algorithm
	if !utils.IsNumeric(req.Order) || !utils.ValidateLuhn(req.Order) {
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	// Process withdrawal
	ctx := r.Context()
	err := h.Repo.WithdrawBalance(ctx, userID, req.Order, req.Sum)
	if err != nil {
		if err.Error() == "insufficient funds" {
			http.Error(w, "Insufficient funds", http.StatusPaymentRequired)
			return
		}
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetWithdrawals returns the list of user's withdrawals
func (h *Handler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get withdrawals
	ctx := r.Context()
	withdrawals, err := h.Repo.GetUserWithdrawals(ctx, userID)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// If no withdrawals, return 204
	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Prepare response
	type withdrawalResponse struct {
		Order       string    `json:"order"`
		Sum         float64   `json:"sum"`
		ProcessedAt time.Time `json:"processed_at"`
	}

	response := make([]withdrawalResponse, 0, len(withdrawals))
	for _, w := range withdrawals {
		response = append(response, withdrawalResponse{
			Order:       w.Order,
			Sum:         w.Sum,
			ProcessedAt: w.ProcessedAt,
		})
	}

	// Return withdrawals
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// processOrder processes an order in background
func (h *Handler) processOrder(orderNumber string) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Set initial status
	err := h.Repo.UpdateOrderStatus(ctx, orderNumber, models.StatusProcessing, 0)
	if err != nil {
		return
	}

	// Get accrual from external service
	accrualResp, err := h.AccrualSvc.GetOrderAccrual(ctx, orderNumber)
	if err != nil || accrualResp == nil {
		// If error or nil response, try again later (in real implementation)
		return
	}

	// If status is final, update order
	if accrualResp.Status == models.StatusProcessed || accrualResp.Status == models.StatusInvalid {
		h.Repo.UpdateOrderStatus(ctx, orderNumber, accrualResp.Status, accrualResp.Accrual)
	}
}
