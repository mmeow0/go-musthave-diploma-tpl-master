package util

import "strconv"

func IsValidLuhn(number string) bool {
	if number == "" {
		return false
	}

	sum := 0
	parity := len(number) % 2

	for i, digit := range number {
		d, err := strconv.Atoi(string(digit))
		if err != nil {
			return false
		}

		if i%2 == parity {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}

		sum += d
	}

	return sum%10 == 0
}
