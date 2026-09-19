package converter

import (
	"errors"
	"fmt"
	"time"
)

type TimeResolution string

const (
	ResolutionNormal      TimeResolution = "normal"
	ResolutionAmbiguous   TimeResolution = "ambiguous"
	ResolutionNonExistent TimeResolution = "non-existent"
)

// TimeConversionRequest holds the parameters for a time conversion.
type TimeConversionRequest struct {
	Year   int `json:"year"`
	Month  int `json:"month"`
	Day    int `json:"day"`
	Hour   int `json:"hour"`
	Minute int `json:"minute"`
	Second int `json:"second"`

	SourceTZ string `json:"source_tz"` // e.g., "America/New_York"
	DestTZ   string `json:"dest_tz"`   // e.g., "Asia/Tokyo"

	// Policy for ambiguous times (fall back): "first" (default) or "second"
	AmbiguousPolicy string `json:"ambiguous_policy"`
	// Policy for non-existent times (spring forward): "forward" (default, shifts forward) or "error"
	NonExistentPolicy string `json:"non_existent_policy"`
}

// TimeConversionResponse holds the result of a time conversion.
type TimeConversionResponse struct {
	Resolution TimeResolution `json:"resolution"`

	SourceTimeUTC   string `json:"source_time_utc"`   // ISO 8601
	SourceTimeLocal string `json:"source_time_local"` // ISO 8601 with offset
	SourceOffset    string `json:"source_offset"`     // e.g. "-04:00"
	SourceZoneName  string `json:"source_zone_name"`  // e.g. "EDT"

	DestTimeUTC   string `json:"dest_time_utc"`   // ISO 8601
	DestTimeLocal string `json:"dest_time_local"` // ISO 8601 with offset
	DestOffset    string `json:"dest_offset"`
	DestZoneName  string `json:"dest_zone_name"`

	IsNextDay bool `json:"is_next_day"`
	IsPrevDay bool `json:"is_prev_day"`

	Warning string `json:"warning,omitempty"`
}

// resolveLocalTime determines the exact UTC instant of a requested local time, checking for ambiguity and non-existence.
func resolveLocalTime(req TimeConversionRequest, loc *time.Location) (time.Time, TimeResolution, string, error) {
	t := time.Date(req.Year, time.Month(req.Month), req.Day, req.Hour, req.Minute, req.Second, 0, loc)

	// Check non-existent
	if t.Year() != req.Year || int(t.Month()) != req.Month || t.Day() != req.Day ||
		t.Hour() != req.Hour || t.Minute() != req.Minute || t.Second() != req.Second {

		if req.NonExistentPolicy == "error" {
			return t, ResolutionNonExistent, "", fmt.Errorf("local time %04d-%02d-%02d %02d:%02d:%02d is non-existent in timezone %s",
				req.Year, req.Month, req.Day, req.Hour, req.Minute, req.Second, req.SourceTZ)
		}

		warning := fmt.Sprintf("The requested local time does not exist in %s due to a daylight saving time gap. It has been automatically shifted forward to %s.",
			req.SourceTZ, t.Format("15:04:05 MST"))
		return t, ResolutionNonExistent, warning, nil
	}

	// Check ambiguous
	ambiguous := false
	var time2 time.Time

	// Scan 4 hours around the time in 1-minute steps to find if any other UTC instant maps to the same local time
	for d := -4 * 60; d <= 4*60; d++ {
		if d == 0 {
			continue
		}
		testT := t.UTC().Add(time.Duration(d) * time.Minute).In(loc)
		if testT.Year() == req.Year && int(testT.Month()) == req.Month && testT.Day() == req.Day &&
			testT.Hour() == req.Hour && testT.Minute() == req.Minute && testT.Second() == req.Second {
			ambiguous = true
			time2 = testT
			break
		}
	}

	if ambiguous {
		// time.Date returns the first occurrence (smaller UTC time)
		// Ensure t1 is the first occurrence and t2 is the second
		t1 := t
		t2 := time2
		if t2.UTC().Before(t1.UTC()) {
			t1, t2 = t2, t1
		}

		warning := fmt.Sprintf("The requested local time is ambiguous in %s (e.g. during a daylight saving time fallback).", req.SourceTZ)

		if req.AmbiguousPolicy == "second" {
			return t2, ResolutionAmbiguous, warning + " The second occurrence was selected.", nil
		}
		return t1, ResolutionAmbiguous, warning + " The first occurrence was selected by default.", nil
	}

	return t, ResolutionNormal, "", nil
}

func formatOffset(offsetSeconds int) string {
	sign := "+"
	if offsetSeconds < 0 {
		sign = "-"
		offsetSeconds = -offsetSeconds
	}
	hours := offsetSeconds / 3600
	minutes := (offsetSeconds % 3600) / 60
	return fmt.Sprintf("%s%02d:%02d", sign, hours, minutes)
}

// ConvertTime performs the global time conversion.
func ConvertTime(req TimeConversionRequest) (*TimeConversionResponse, error) {
	if req.SourceTZ == "" || req.DestTZ == "" {
		return nil, errors.New("source_tz and dest_tz are required")
	}

	sourceLoc, err := time.LoadLocation(req.SourceTZ)
	if err != nil {
		return nil, fmt.Errorf("invalid source timezone: %s", req.SourceTZ)
	}

	destLoc, err := time.LoadLocation(req.DestTZ)
	if err != nil {
		return nil, fmt.Errorf("invalid destination timezone: %s", req.DestTZ)
	}

	// 1-4. Resolve the local time into an exact global instant (UTC)
	sourceTime, resolution, warning, err := resolveLocalTime(req, sourceLoc)
	if err != nil {
		return nil, err
	}

	// 5-6. Apply destination timezone rules
	destTime := sourceTime.In(destLoc)

	// Format zone names and offsets
	srcZone, srcOffsetSecs := sourceTime.Zone()
	dstZone, dstOffsetSecs := destTime.Zone()

	// Determine date change
	// Create normalized dates (without time)
	srcDate := time.Date(sourceTime.Year(), sourceTime.Month(), sourceTime.Day(), 0, 0, 0, 0, time.UTC)
	dstDate := time.Date(destTime.Year(), destTime.Month(), destTime.Day(), 0, 0, 0, 0, time.UTC)

	isNextDay := dstDate.After(srcDate)
	isPrevDay := dstDate.Before(srcDate)

	res := &TimeConversionResponse{
		Resolution:      resolution,
		SourceTimeUTC:   sourceTime.UTC().Format(time.RFC3339),
		SourceTimeLocal: sourceTime.Format(time.RFC3339),
		SourceOffset:    formatOffset(srcOffsetSecs),
		SourceZoneName:  srcZone,

		DestTimeUTC:   destTime.UTC().Format(time.RFC3339),
		DestTimeLocal: destTime.Format(time.RFC3339),
		DestOffset:    formatOffset(dstOffsetSecs),
		DestZoneName:  dstZone,

		IsNextDay: isNextDay,
		IsPrevDay: isPrevDay,
		Warning:   warning,
	}

	return res, nil
}
