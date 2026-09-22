package phone

import (
	"errors"
	"strings"
	"unicode"
)

var ErrInvalidNumber = errors.New("invalid phone number")

// Normalize accepts an international number and returns canonical E.164-like digits.
// V1 deliberately requires an explicit calling code. National-format parsing will be
// introduced with country metadata instead of guessing a user's country.
func Normalize(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" { return "", ErrInvalidNumber }

	var b strings.Builder
	for i, r := range input {
		switch {
		case unicode.IsDigit(r):
			b.WriteRune(r)
		case r == '+' && i == 0:
		case r == ' ', r == '-', r == '(', r == ')', r == '.':
		default:
			return "", ErrInvalidNumber
		}
	}

	digits := b.String()
	if !strings.HasPrefix(input, "+") || len(digits) < 8 || len(digits) > 15 || digits[0] == '0' {
		return "", ErrInvalidNumber
	}
	return "+" + digits, nil
}
