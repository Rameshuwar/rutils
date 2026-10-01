package nifty

import (
	"context"
	"time"
)

// Service provides high-level business operations for NIFTY 50 data
type Service struct {
	storage Storage
	client  NSEClient
	syncer  *Syncer
}

// NewService creates a new Service instance
func NewService(storage Storage, client NSEClient) *Service {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		loc = time.FixedZone("IST", 5*3600+30*60)
	}

	syncer := NewSyncer(client, storage, loc)

	return &Service{
		storage: storage,
		client:  client,
		syncer:  syncer,
	}
}

// GetCompanies returns the latest stored NIFTY 50 constituent companies
func (s *Service) GetCompanies() (*NiftyData, error) {
	return s.storage.Load()
}

// Sync downloads the latest constituent data from NSE and stores it safely
func (s *Service) Sync(ctx context.Context) (*NiftyData, error) {
	return s.syncer.Sync(ctx)
}

// StartScheduler initializes and starts the in-process background cron scheduler
func (s *Service) StartScheduler(ctx context.Context) *Scheduler {
	scheduler := NewScheduler(s)
	scheduler.Start(ctx)
	return scheduler
}
