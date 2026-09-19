package converter

import (
	"errors"
)

type RailwayConversionRequest struct {
	Direction string `json:"direction"` // "12to24" or "24to12"
	Hour      int    `json:"hour"`
	Minute    int    `json:"minute"`
	AmPm      string `json:"ampm,omitempty"` // "AM" or "PM"
}

type RailwayConversionResponse struct {
	Hour12 int    `json:"hour_12"`
	Hour24 int    `json:"hour_24"`
	Minute int    `json:"minute"`
	AmPm   string `json:"ampm"`
}

func ConvertRailwayTime(req RailwayConversionRequest) (*RailwayConversionResponse, error) {
	if req.Minute < 0 || req.Minute > 59 {
		return nil, errors.New("minute must be between 0 and 59")
	}

	res := &RailwayConversionResponse{
		Minute: req.Minute,
	}

	if req.Direction == "12to24" {
		if req.Hour < 1 || req.Hour > 12 {
			return nil, errors.New("12-hour format hour must be between 1 and 12")
		}
		if req.AmPm != "AM" && req.AmPm != "PM" {
			return nil, errors.New("ampm must be 'AM' or 'PM'")
		}

		res.Hour12 = req.Hour
		res.AmPm = req.AmPm

		h24 := req.Hour
		if req.AmPm == "PM" && h24 != 12 {
			h24 += 12
		}
		if req.AmPm == "AM" && h24 == 12 {
			h24 = 0
		}
		res.Hour24 = h24

	} else if req.Direction == "24to12" {
		if req.Hour < 0 || req.Hour > 23 {
			return nil, errors.New("24-hour format hour must be between 0 and 23")
		}

		res.Hour24 = req.Hour

		h12 := req.Hour
		ampm := "AM"
		
		if h12 >= 12 {
			ampm = "PM"
		}
		
		h12 = h12 % 12
		if h12 == 0 {
			h12 = 12
		}
		
		res.Hour12 = h12
		res.AmPm = ampm

	} else {
		return nil, errors.New("direction must be '12to24' or '24to12'")
	}

	return res, nil
}
