package helpers

import "math"

//Round2 rounds to 2 decimal places
func Round2(value float64) float64 {
	return math.Round(value*100) / 100
}
