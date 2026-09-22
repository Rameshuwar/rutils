package converter

import (
	"fmt"
)

// CalculateBMI calculates the Body Mass Index and returns the value and its category.
// It accepts weight and height with their respective units, converting them to kg and meters.
func CalculateBMI(weight float64, weightUnit string, height float64, heightUnit string) (float64, string, error) {
	// Convert weight to kilograms
	weightKg, err := ConvertMeasurement("weight", weightUnit, "kilograms", weight)
	if err != nil {
		return 0, "", fmt.Errorf("failed to convert weight to kg: %v", err)
	}

	// Convert height to meters
	heightM, err := ConvertMeasurement("length", heightUnit, "meters", height)
	if err != nil {
		return 0, "", fmt.Errorf("failed to convert height to meters: %v", err)
	}

	if heightM <= 0 {
		return 0, "", fmt.Errorf("height must be greater than zero")
	}
	if weightKg <= 0 {
		return 0, "", fmt.Errorf("weight must be greater than zero")
	}

	bmi := weightKg / (heightM * heightM)
	category := getBMICategory(bmi)

	return bmi, category, nil
}

func getBMICategory(bmi float64) string {
	switch {
	case bmi < 18.5:
		return "Underweight"
	case bmi >= 18.5 && bmi < 25:
		return "Normal weight"
	case bmi >= 25 && bmi < 30:
		return "Overweight"
	default:
		return "Obesity"
	}
}
