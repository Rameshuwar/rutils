package converter

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func getBaseInt(base string) (int, error) {
	switch strings.ToLower(base) {
	case "binary":
		return 2, nil
	case "octal":
		return 8, nil
	case "decimal":
		return 10, nil
	case "hexadecimal":
		return 16, nil
	default:
		return 0, fmt.Errorf("unsupported numeral base: %s", base)
	}
}

// ConvertNumeral converts a string value from one numeral system to another.
func ConvertNumeral(fromBase, toBase, value string) (string, error) {
	if value == "" {
		return "", errors.New("value cannot be empty")
	}

	from, err := getBaseInt(fromBase)
	if err != nil {
		return "", err
	}

	to, err := getBaseInt(toBase)
	if err != nil {
		return "", err
	}

	// Parse string into integer based on fromBase
	parsedInt, err := strconv.ParseInt(value, from, 64)
	if err != nil {
		return "", fmt.Errorf("invalid value '%s' for base %s", value, fromBase)
	}

	// Format integer back to string based on toBase
	result := strconv.FormatInt(parsedInt, to)

	// Return uppercase for hexadecimal (e.g., 1a -> 1A)
	if to == 16 {
		result = strings.ToUpper(result)
	}

	return result, nil
}
