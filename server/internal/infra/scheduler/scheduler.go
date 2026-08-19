package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/rmf87/divoene/internal/core/services"
)

// Scheduler manages background periodic tasks.
type Scheduler struct {
	contractSvc *services.ContractService
	paymentSvc  *services.PaymentService
}

// NewScheduler creates a new scheduler.
func NewScheduler(contractSvc *services.ContractService, paymentSvc *services.PaymentService) *Scheduler {
	return &Scheduler{
		contractSvc: contractSvc,
		paymentSvc:  paymentSvc,
	}
}

// Start begins all background goroutines. Blocks until ctx is cancelled.
func (s *Scheduler) Start(ctx context.Context) {
	// Hourly contract poll — was Cloud Scheduler
	go func() {
		s.pollContracts(ctx)
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.pollContracts(ctx)
			}
		}
	}()

	// Daily guide payout at 20:00 BRT — was Cloud Scheduler
	go func() {
		for {
			now := time.Now()
			loc, _ := time.LoadLocation("America/Sao_Paulo")
			if loc != nil {
				now = now.In(loc)
			}
			next := time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, now.Location())
			if now.After(next) {
				next = next.Add(24 * time.Hour)
			}
			time.Sleep(time.Until(next))

			select {
			case <-ctx.Done():
				return
			default:
				s.processDailyPayout(ctx)
			}
		}
	}()
}

func (s *Scheduler) pollContracts(ctx context.Context) {
	log.Printf("[scheduler] polling contracts...")
	// Contract polling is handled by Clicksign webhooks.
	// This is kept as a placeholder for future reconciliation.
}

func (s *Scheduler) processDailyPayout(ctx context.Context) {
	log.Printf("[scheduler] processing daily guide payout...")
	// Guide payout logic — placeholder for future implementation.
}
