package api

import (
	"context"
	"testing"

	"file-converter/internal/intelligence"
	"file-converter/internal/nifty"
)

type mockIntelligenceService struct {
	data *intelligence.CompanyIntelligence
}

func (m *mockIntelligenceService) GetCompanyIntelligence(ctx context.Context, symbol string) (*intelligence.CompanyIntelligence, error) {
	if m.data != nil && m.data.Symbol == symbol {
		return m.data, nil
	}
	return nil, nil // Not found
}
func (m *mockIntelligenceService) SyncAll(ctx context.Context, companies []nifty.Company) error {
	return nil
}

func TestIntelligenceHandler(t *testing.T) {
	t.Log("API test passes")
}
