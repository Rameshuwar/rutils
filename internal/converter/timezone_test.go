package converter

import (
	"testing"
)

func TestTimezoneConversion_Basic(t *testing.T) {
	// Standard offset check (Tokyo is +09:00 always)
	req := TimeConversionRequest{
		Year:     2024,
		Month:    6,
		Day:      1,
		Hour:     12,
		Minute:   0,
		Second:   0,
		SourceTZ: "Asia/Tokyo",
		DestTZ:   "UTC",
	}

	res, err := ConvertTime(req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if res.Resolution != ResolutionNormal {
		t.Errorf("Expected normal resolution, got %v", res.Resolution)
	}
	if res.SourceOffset != "+09:00" {
		t.Errorf("Expected SourceOffset +09:00, got %s", res.SourceOffset)
	}
	if res.DestTimeUTC != "2024-06-01T03:00:00Z" { // 12 - 9 = 3
		t.Errorf("Expected DestTimeUTC 2024-06-01T03:00:00Z, got %s", res.DestTimeUTC)
	}
}

func TestTimezoneConversion_DateRollover(t *testing.T) {
	// Next day rollover
	req := TimeConversionRequest{
		Year:     2024,
		Month:    12,
		Day:      31,
		Hour:     20,
		Minute:   0,
		Second:   0,
		SourceTZ: "America/New_York",
		DestTZ:   "Asia/Tokyo",
	}

	res, err := ConvertTime(req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// NY is UTC-5 in Dec. Tokyo is UTC+9. 20:00 NY -> 01:00 UTC next day -> 10:00 Tokyo next day.
	if !res.IsNextDay {
		t.Errorf("Expected IsNextDay to be true")
	}
	if res.DestTimeLocal != "2025-01-01T10:00:00+09:00" {
		t.Errorf("Expected 2025-01-01T10:00:00+09:00, got %s", res.DestTimeLocal)
	}
}

func TestTimezoneConversion_NonExistent(t *testing.T) {
	// Spring forward gap in New York: March 10, 2024 02:00 -> 03:00
	req := TimeConversionRequest{
		Year:     2024,
		Month:    3,
		Day:      10,
		Hour:     2,
		Minute:   30,
		Second:   0,
		SourceTZ: "America/New_York",
		DestTZ:   "UTC",
	}

	res, err := ConvertTime(req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if res.Resolution != ResolutionNonExistent {
		t.Errorf("Expected non-existent resolution, got %v", res.Resolution)
	}

	// Error policy test
	req.NonExistentPolicy = "error"
	_, err = ConvertTime(req)
	if err == nil {
		t.Errorf("Expected error for non-existent time with error policy")
	}
}

func TestTimezoneConversion_Ambiguous(t *testing.T) {
	// Fall back overlap in New York: November 3, 2024 02:00 -> 01:00
	req := TimeConversionRequest{
		Year:     2024,
		Month:    11,
		Day:      3,
		Hour:     1,
		Minute:   30,
		Second:   0,
		SourceTZ: "America/New_York",
		DestTZ:   "UTC",
	}

	res, err := ConvertTime(req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if res.Resolution != ResolutionAmbiguous {
		t.Errorf("Expected ambiguous resolution, got %v", res.Resolution)
	}

	// Default is first occurrence (EDT, UTC-4)
	if res.SourceOffset != "-04:00" {
		t.Errorf("Expected offset -04:00, got %s", res.SourceOffset)
	}

	// Request second occurrence
	req.AmbiguousPolicy = "second"
	res2, _ := ConvertTime(req)
	if res2.SourceOffset != "-05:00" {
		t.Errorf("Expected offset -05:00 for second occurrence, got %s", res2.SourceOffset)
	}
}

func TestTimezoneConversion_UnusualOffsets(t *testing.T) {
	// Nepal is +05:45
	req := TimeConversionRequest{
		Year:     2024,
		Month:    1,
		Day:      1,
		Hour:     12,
		Minute:   0,
		Second:   0,
		SourceTZ: "Asia/Kathmandu",
		DestTZ:   "UTC",
	}

	res, err := ConvertTime(req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if res.SourceOffset != "+05:45" {
		t.Errorf("Expected offset +05:45, got %s", res.SourceOffset)
	}
	if res.DestTimeUTC != "2024-01-01T06:15:00Z" {
		t.Errorf("Expected 2024-01-01T06:15:00Z, got %s", res.DestTimeUTC)
	}
}

func TestTimezoneConversion_Historical(t *testing.T) {
	// Historical offset test - e.g. Paris in 1900 had a weird offset (+00:09)
	req := TimeConversionRequest{
		Year:     1900,
		Month:    1,
		Day:      1,
		Hour:     12,
		Minute:   0,
		Second:   0,
		SourceTZ: "Europe/Paris",
		DestTZ:   "UTC",
	}

	res, err := ConvertTime(req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if res.SourceOffset != "+00:09" {
		t.Errorf("Expected offset +00:09, got %s", res.SourceOffset)
	}
}
