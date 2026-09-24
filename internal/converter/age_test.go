package converter

import "testing"

func TestCalculateAgeData_ExactAgeAndNextBirthday(t *testing.T) {
	res, err := CalculateAgeData("1990-05-15", "2025-09-22")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Age.Years != 35 || res.Age.Months != 4 || res.Age.Days != 7 {
		t.Fatalf("unexpected age result: %+v", res.Age)
	}

	if res.NextBirthday.Date != "2026-05-15" {
		t.Fatalf("expected next birthday date 2026-05-15, got %s", res.NextBirthday.Date)
	}
	if res.NextBirthday.Month != 5 || res.NextBirthday.Day != 15 {
		t.Fatalf("expected next birthday month/day to be 5/15, got %d/%d", res.NextBirthday.Month, res.NextBirthday.Day)
	}
	if res.NextBirthday.DayOfWeek == "" {
		t.Fatal("expected next birthday weekday to be populated")
	}
}
