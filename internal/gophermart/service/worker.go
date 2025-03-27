package service

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/25x8/ya-prakt-6-sprint/internal/gophermart/models"
	"github.com/25x8/ya-prakt-6-sprint/internal/gophermart/repository"
)

// OrderProcessor processes orders in the background
type OrderProcessor struct {
	repo       repository.Repository
	accrualSvc *AccrualService
	interval   time.Duration
	stopCh     chan struct{}
	wg         sync.WaitGroup
}

// NewOrderProcessor creates a new order processor
func NewOrderProcessor(repo repository.Repository, accrualSvc *AccrualService) *OrderProcessor {
	return &OrderProcessor{
		repo:       repo,
		accrualSvc: accrualSvc,
		interval:   5 * time.Second, // Check for new orders every 5 seconds
		stopCh:     make(chan struct{}),
	}
}

// Start starts the order processor
func (p *OrderProcessor) Start() {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.processLoop()
	}()
}

// Stop stops the order processor
func (p *OrderProcessor) Stop() {
	close(p.stopCh)
	p.wg.Wait()
}

// processLoop is the main processing loop
func (p *OrderProcessor) processLoop() {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Process pending orders
			p.processPendingOrders()
		case <-p.stopCh:
			return
		}
	}
}

// processPendingOrders processes all pending orders
func (p *OrderProcessor) processPendingOrders() {
	// In a real implementation, we would query for pending orders
	// and process them in batches.
	// For simplicity, this implementation is just a placeholder.
}

// processOrder processes a single order
func (p *OrderProcessor) processOrder(ctx context.Context, order *models.Order) {
	// Skip already processed orders
	if order.Status == models.StatusProcessed || order.Status == models.StatusInvalid {
		return
	}

	// Update status to PROCESSING if it's NEW
	if order.Status == models.StatusNew {
		if err := p.repo.UpdateOrderStatus(ctx, order.Number, models.StatusProcessing, 0); err != nil {
			log.Printf("Error updating order %s status: %v", order.Number, err)
			return
		}
	}

	// Get accrual information
	accrualResp, err := p.accrualSvc.GetOrderAccrual(ctx, order.Number)
	if err != nil {
		log.Printf("Error getting accrual for order %s: %v", order.Number, err)
		return
	}

	// If no response or not final status, skip for now
	if accrualResp == nil || (accrualResp.Status != models.StatusProcessed && accrualResp.Status != models.StatusInvalid) {
		return
	}

	// Update order with final status
	if err := p.repo.UpdateOrderStatus(ctx, order.Number, accrualResp.Status, accrualResp.Accrual); err != nil {
		log.Printf("Error updating order %s with accrual: %v", order.Number, err)
	}
}
