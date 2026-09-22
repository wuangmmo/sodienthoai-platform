package phone

import (
	"errors"
	"strings"
)

var ErrInvalidNumber = errors.New("invalid phone number")

func Normalize(input string) (string, error) { return NormalizeForCountry(input, "") }

// NormalizeForCountry converts supported user input to canonical E.164.
// Only ASCII digits are accepted because E.164 is an ASCII numeric identifier.
// National-format parsing is intentionally country-scoped to avoid ambiguous
// global guesses.
func NormalizeForCountry(input, country string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", ErrInvalidNumber
	}

	international := strings.HasPrefix(input, "+")
	var b strings.Builder
	for i, r := range input {
		switch {
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '+' && i == 0:
			// Leading international marker; it is restored after validation.
		case r == ' ' || r == '-' || r == '(' || r == ')' || r == '.':
			// Common presentation characters are ignored.
		default:
			return "", ErrInvalidNumber
		}
	}

	digits := b.String()
	if international {
		if len(digits) < 8 || len(digits) > 15 || digits[0] == '0' {
			return "", ErrInvalidNumber
		}
		return "+" + digits, nil
	}

	if strings.EqualFold(strings.TrimSpace(country), "VN") {
		// Vietnamese national numbers must start with the trunk prefix 0.
		// Keep validation deliberately structural here; allocation/prefix
		// intelligence belongs in metadata, not canonicalization.
		if len(digits) < 9 || len(digits) > 11 || digits[0] != '0' {
			return "", ErrInvalidNumber
		}
		national := strings.TrimPrefix(digits, "0")
		if national == "" || national[0] == '0' {
			return "", ErrInvalidNumber
		}
		e164 := "+84" + national
		if len(e164)-1 > 15 {
			return "", ErrInvalidNumber
		}
		return e164, nil
	}

	return "", ErrInvalidNumber
}
