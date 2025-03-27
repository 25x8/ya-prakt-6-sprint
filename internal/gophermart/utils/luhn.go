package utils

import (
	"strconv"
)

// ValidateLuhn checks if a string passes the Luhn algorithm check
func ValidateLuhn(number string) bool {
	// Convert string to slice of digits
	digits := make([]int, 0, len(number))
	for _, r := range number {
		if r < '0' || r > '9' {
			return false // Non-digit character
		}
		digits = append(digits, int(r-'0'))
	}

	// Luhn algorithm
	sum := 0
	parity := len(digits) % 2
	for i, digit := range digits {
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}

	return sum%10 == 0
}

// IsNumeric checks if a string contains only digits
func IsNumeric(s string) bool {
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil
}
