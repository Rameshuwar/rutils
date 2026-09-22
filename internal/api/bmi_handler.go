package api

import (
	"encoding/json"
	"net/http"

	"file-converter/internal/converter"
)

type BMIRequest struct {
	Weight     float64 `json:"weight"`
	WeightUnit string  `json:"weightUnit"` // e.g., "kilograms", "pounds"
	Height     float64 `json:"height"`
	HeightUnit string  `json:"heightUnit"` // e.g., "meters", "centimeters", "feet", "inches"
}

type BMIResponse struct {
	BMI      float64 `json:"bmi"`
	Category string  `json:"category"`
}

// HandleBMICalculate handles HTTP requests for BMI calculations
// @Summary BMI Calculator
// @Description Calculates Body Mass Index (BMI) based on weight and height
// @Accept json
// @Produce json
// @Param request body api.BMIRequest true "BMI Calculation Request"
// @Success 200 {object} api.BMIResponse "BMI Result"
// @Failure 400 {string} string "Bad Request"
// @Router /calculate-bmi [post]
func HandleBMICalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BMIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.WeightUnit == "" || req.HeightUnit == "" {
		http.Error(w, "weightUnit and heightUnit are required", http.StatusBadRequest)
		return
	}

	bmi, category, err := converter.CalculateBMI(req.Weight, req.WeightUnit, req.Height, req.HeightUnit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res := BMIResponse{
		BMI:      bmi,
		Category: category,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
