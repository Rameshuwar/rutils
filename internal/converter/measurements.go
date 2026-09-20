package converter

import (
	"fmt"
	"strings"
)

// ConversionFactors stores multipliers to convert a unit to its category's base unit.
// e.g. for length (base: meter), kilometer is 1000, centimeter is 0.01.
var ConversionFactors = map[string]map[string]float64{
	"length": {
		"meters":      1.0,
		"kilometers":  1000.0,
		"centimeters": 0.01,
		"millimeters": 0.001,
		"miles":       1609.344,
		"yards":       0.9144,
		"feet":        0.3048,
		"inches":      0.0254,
	},
	"weight": {
		"kilograms":  1.0, // base unit
		"grams":      0.001,
		"milligrams": 0.000001,
		"pounds":     0.45359237,
		"ounces":     0.028349523125,
	},
	"volume": {
		"liters":       1.0, // base unit
		"milliliters":  0.001,
		"gallons":      3.78541,
		"quarts":       0.946353,
		"pints":        0.473176,
		"fluid_ounces": 0.0295735,
	},
	"area": {
		"square_meters":     1.0, // base unit
		"square_kilometers": 1000000.0,
		"hectares":          10000.0,
		"acres":             4046.8564224,
		"square_feet":       0.09290304,
		"square_miles":      2589988.11,
	},
	"time": {
		"seconds": 1.0, // base unit
		"minutes": 60.0,
		"hours":   3600.0,
		"days":    86400.0,
		"weeks":   604800.0,
	},
	"speed": {
		"meters_per_second":   1.0,             // base unit
		"kilometers_per_hour": 1000.0 / 3600.0, // ~0.277778
		"miles_per_hour":      0.44704,
		"feet_per_second":     0.3048,
		"knots":               0.514444,
	},
	"data": {
		"bytes":     1.0, // base unit
		"kilobytes": 1024.0,
		"megabytes": 1048576.0,          // 1024^2
		"gigabytes": 1073741824.0,       // 1024^3
		"terabytes": 1099511627776.0,    // 1024^4
		"petabytes": 1125899906842624.0, // 1024^5
		"bits":      0.125,              // 1 byte = 8 bits
	},
}

// ConvertMeasurement handles the conversion logic
func ConvertMeasurement(category, fromUnit, toUnit string, value float64) (float64, error) {
	category = strings.ToLower(category)
	fromUnit = strings.ToLower(fromUnit)
	toUnit = strings.ToLower(toUnit)

	// Handle temperature as a special case
	if category == "temperature" {
		return convertTemperature(fromUnit, toUnit, value)
	}

	// Lookup category
	factors, ok := ConversionFactors[category]
	if !ok {
		return 0, fmt.Errorf("unsupported category: %s", category)
	}

	// Lookup units
	fromFactor, fromOk := factors[fromUnit]
	if !fromOk {
		return 0, fmt.Errorf("unsupported fromUnit: %s in category: %s", fromUnit, category)
	}

	toFactor, toOk := factors[toUnit]
	if !toOk {
		return 0, fmt.Errorf("unsupported toUnit: %s in category: %s", toUnit, category)
	}

	// Convert: Value -> Base Unit -> Target Unit
	baseValue := value * fromFactor
	result := baseValue / toFactor

	return result, nil
}

func convertTemperature(fromUnit, toUnit string, value float64) (float64, error) {
	// Convert fromUnit to Celsius (Base)
	var celsiusValue float64
	switch fromUnit {
	case "celsius":
		celsiusValue = value
	case "fahrenheit":
		celsiusValue = (value - 32) * 5 / 9
	case "kelvin":
		celsiusValue = value - 273.15
	default:
		return 0, fmt.Errorf("unsupported temperature fromUnit: %s", fromUnit)
	}

	// Convert Celsius to toUnit
	switch toUnit {
	case "celsius":
		return celsiusValue, nil
	case "fahrenheit":
		return (celsiusValue * 9 / 5) + 32, nil
	case "kelvin":
		return celsiusValue + 273.15, nil
	default:
		return 0, fmt.Errorf("unsupported temperature toUnit: %s", toUnit)
	}
}
