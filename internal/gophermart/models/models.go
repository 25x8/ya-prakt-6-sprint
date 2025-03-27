package models

import (
	"time"
)

// User represents a registered user
type User struct {
	ID           int64     `json:"id"`
	Login        string    `json:"login"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// Order represents an order in the system
type Order struct {
	ID         int64     `json:"id"`
	Number     string    `json:"number"`
	UserID     int64     `json:"user_id"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// Balance represents a user's loyalty balance
type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// Withdrawal represents a withdrawal transaction
type Withdrawal struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

// AccrualResponse represents the response from the accrual system
type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

// OrderStatuses constants for order status
const (
	StatusNew        = "NEW"
	StatusProcessing = "PROCESSING"
	StatusInvalid    = "INVALID"
	StatusProcessed  = "PROCESSED"
)

// AccrualStatuses constants for accrual system statuses
const (
	StatusRegistered = "REGISTERED"
)
