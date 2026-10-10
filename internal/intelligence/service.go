package intelligence

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"file-converter/internal/nifty"
)

type Service struct {
	storage Storage
	fetcher Fetcher
	niftySvc *nifty.Service
}

func NewService(storage Storage, fetcher Fetcher, niftySvc *nifty.Service) *Service {
	return &Service{
		storage:  storage,
		fetcher:  fetcher,
		niftySvc: niftySvc,
	}
}

// GetIntelligence retrieves the intelligence data for a given symbol
func (s *Service) GetIntelligence(symbol string) (*CompanyIntelligence, error) {
	return s.storage.Load(symbol)
}

// SyncAll syncs all active companies in Nifty 50.
func (s *Service) SyncAll(ctx context.Context) error {
	niftyData, err := s.niftySvc.GetCompanies()
	if err != nil {
		return err
	}

	activeSymbols := make(map[string]bool)
	for _, company := range niftyData.Companies {
		activeSymbols[company.Symbol] = true
	}

	// Process sequentially with a delay to avoid rate-limiting (HTTP 429) from data sources
	for _, company := range niftyData.Companies {
		time.Sleep(2 * time.Second)
		
		ci, err := s.fetcher.Fetch(ctx, company)
		if err != nil {
			log.Printf("[INTELLIGENCE] Failed to fetch data for %s: %v", company.Symbol, err)
			// Create a minimal fallback if doesn't exist
			if _, loadErr := s.storage.Load(company.Symbol); loadErr != nil {
				fallback := &CompanyIntelligence{
					Symbol:      company.Symbol,
					CompanyName: company.CompanyName,
					Industry:    company.Industry,
					FetchStatus: "failed",
					UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
				}
				_ = s.storage.Save(company.Symbol, fallback)
			}
			continue
		}
		
		// Load old to preserve custom fields if needed, but here we just replace
		if err := s.storage.Save(company.Symbol, ci); err != nil {
			log.Printf("[INTELLIGENCE] Failed to save data for %s: %v", company.Symbol, err)
		}
	}

	// Cleanup removed companies
	if fs, ok := s.storage.(*FileStorage); ok {
		entries, err := os.ReadDir(fs.baseDir)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
					symbol := strings.TrimSuffix(entry.Name(), ".json")
					if !activeSymbols[symbol] {
						log.Printf("[INTELLIGENCE] Removing data for %s as it is no longer in NIFTY 50", symbol)
						_ = s.storage.Delete(symbol)
					}
				}
			}
		}
	}

	log.Printf("[INTELLIGENCE] Synced all companies")

	return nil
}

func (s *Service) StartScheduler(ctx context.Context) {
	go func() {
		// Run initial sync after a short delay
		select {
		case <-time.After(10 * time.Second):
			s.SyncAll(ctx)
		case <-ctx.Done():
			return
		}

		// Then run daily
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.SyncAll(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}
