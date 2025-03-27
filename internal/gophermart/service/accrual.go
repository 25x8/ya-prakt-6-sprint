package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/25x8/ya-prakt-6-sprint/internal/gophermart/models"
)

// AccrualService handles communication with the accrual service
type AccrualService struct {
	baseURL    string
	httpClient *http.Client
}

// NewAccrualService creates a new accrual service
func NewAccrualService(baseURL string) *AccrualService {
	return &AccrualService{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetOrderAccrual fetches the accrual information for an order
func (s *AccrualService) GetOrderAccrual(ctx context.Context, orderNumber string) (*models.AccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", s.baseURL, orderNumber)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Handle rate limiting
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter != "" {
			seconds, err := strconv.Atoi(retryAfter)
			if err == nil {
				return nil, fmt.Errorf("rate limited, retry after %d seconds", seconds)
			}
		}
		return nil, fmt.Errorf("rate limited")
	}

	// Handle 204 No Content (order not registered)
	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	// Handle other errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("accrual service returned status %d", resp.StatusCode)
	}

	// Parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var accrualResp models.AccrualResponse
	if err := json.Unmarshal(body, &accrualResp); err != nil {
		return nil, err
	}

	return &accrualResp, nil
}

// ProcessOrderAccrual processes an order through the accrual system
// and updates its status and accrual in the database
func (s *AccrualService) ProcessOrderAccrual(ctx context.Context, orderNumber string) error {
	// This would be called by a background worker
	// For simplicity, we're not implementing the full background worker here
	return nil
}
