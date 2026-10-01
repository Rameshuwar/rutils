package nifty

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
)

// Syncer handles downloading, validating, and atomically storing NIFTY 50 data
type Syncer struct {
	client   NSEClient
	storage  Storage
	location *time.Location
}

// NewSyncer creates a new Syncer
func NewSyncer(client NSEClient, storage Storage, loc *time.Location) *Syncer {
	if loc == nil {
		var err error
		loc, err = time.LoadLocation("Asia/Kolkata")
		if err != nil {
			loc = time.FixedZone("IST", 5*3600+30*60)
		}
	}
	return &Syncer{
		client:   client,
		storage:  storage,
		location: loc,
	}
}

// Sync performs the download, parse, validation, and safe atomic update
func (s *Syncer) Sync(ctx context.Context) (*NiftyData, error) {
	companies, err := s.client.FetchAndParse(ctx)
	if err != nil {
		if errors.Is(err, ErrInvalidDataset) {
			log.Printf("NIFTY 50 sync failed: invalid or incomplete NSE dataset: %v", err)
			return nil, fmt.Errorf("NIFTY 50 sync failed: invalid or incomplete NSE dataset: %w", err)
		}
		log.Printf("NIFTY 50 sync failed: unable to download NSE data: %v", err)
		return nil, fmt.Errorf("NIFTY 50 sync failed: unable to download NSE data: %w", err)
	}

	if err := ValidateCompanies(companies); err != nil {
		log.Printf("NIFTY 50 sync failed: invalid or incomplete NSE dataset: %v", err)
		return nil, fmt.Errorf("NIFTY 50 sync failed: invalid or incomplete NSE dataset: %w", err)
	}

	now := time.Now().In(s.location)
	niftyData := &NiftyData{
		UpdatedAt: now.Format(time.RFC3339),
		Count:     len(companies),
		Companies: companies,
	}

	if err := s.storage.Save(niftyData); err != nil {
		log.Printf("NIFTY 50 sync failed: unable to save dataset: %v", err)
		return nil, fmt.Errorf("NIFTY 50 sync failed: unable to save dataset: %w", err)
	}

	timeStr := now.Format("2006-01-02 15:04 MST")
	log.Printf("NIFTY 50 sync completed: %d companies updated at %s", len(companies), timeStr)

	return niftyData, nil
}
