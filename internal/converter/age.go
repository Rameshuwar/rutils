package converter

import (
	"fmt"
	"time"
)

type AgeResult struct {
	Years  int `json:"years"`
	Months int `json:"months"`
	Days   int `json:"days"`
}

type NextBirthday struct {
	Date            string `json:"date"`
	DayOfWeek       string `json:"dayOfWeek"`
	Month           int    `json:"month"`
	Day             int    `json:"day"`
	MonthsRemaining int    `json:"monthsRemaining"`
	DaysRemaining   int    `json:"daysRemaining"`
}

type Summary struct {
	Years   int `json:"years"`
	Months  int `json:"months"`
	Weeks   int `json:"weeks"`
	Days    int `json:"days"`
	Hours   int `json:"hours"`
	Minutes int `json:"minutes"`
}

type AgeCalculationResponse struct {
	Age          AgeResult    `json:"age"`
	NextBirthday NextBirthday `json:"nextBirthday"`
	Summary      Summary      `json:"summary"`
}

// calcAgeHelper returns the difference in years, months, and days between two dates.
func calcAgeHelper(start, end time.Time) (y, m, d int) {
	y = end.Year() - start.Year()
	m = int(end.Month()) - int(start.Month())
	d = end.Day() - start.Day()

	if d < 0 {
		m--
		// Get the number of days in the previous month of 'end'
		lastDayOfPrevMonth := time.Date(end.Year(), end.Month(), 0, 0, 0, 0, 0, end.Location())
		d += lastDayOfPrevMonth.Day()
	}

	if m < 0 {
		y--
		m += 12
	}
	return
}

func CalculateAgeData(dobStr, todayStr string) (*AgeCalculationResponse, error) {
	layout := "2006-01-02"
	dob, err := time.Parse(layout, dobStr)
	if err != nil {
		return nil, fmt.Errorf("invalid dob format, expected YYYY-MM-DD")
	}

	today, err := time.Parse(layout, todayStr)
	if err != nil {
		return nil, fmt.Errorf("invalid today format, expected YYYY-MM-DD")
	}

	if dob.After(today) {
		return nil, fmt.Errorf("date of birth cannot be in the future relative to 'today'")
	}

	// 1. Calculate Age
	y, m, d := calcAgeHelper(dob, today)
	ageResult := AgeResult{Years: y, Months: m, Days: d}

	// 2. Calculate Next Birthday
	nextBirthdayThisYear := time.Date(today.Year(), dob.Month(), dob.Day(), 0, 0, 0, 0, today.Location())
	nextBday := nextBirthdayThisYear
	if !today.Before(nextBirthdayThisYear) {
		nextBday = time.Date(today.Year()+1, dob.Month(), dob.Day(), 0, 0, 0, 0, today.Location())
	}

	_, mRemaining, dRemaining := calcAgeHelper(today, nextBday)
	nextBirthdayData := NextBirthday{
		Date:            nextBday.Format(layout),
		DayOfWeek:       nextBday.Weekday().String(),
		Month:           int(nextBday.Month()),
		Day:             nextBday.Day(),
		MonthsRemaining: mRemaining,
		DaysRemaining:   dRemaining,
	}

	// 3. Calculate Summary
	totalDuration := today.Sub(dob)
	totalDays := int(totalDuration.Hours() / 24)
	totalHours := int(totalDuration.Hours())
	totalMinutes := int(totalDuration.Minutes())

	summaryData := Summary{
		Years:   y,
		Months:  (y * 12) + m,
		Weeks:   totalDays / 7,
		Days:    totalDays,
		Hours:   totalHours,
		Minutes: totalMinutes,
	}

	return &AgeCalculationResponse{
		Age:          ageResult,
		NextBirthday: nextBirthdayData,
		Summary:      summaryData,
	}, nil
}
