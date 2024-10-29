package utils

import "math"

// ToFixed rounds the given number to the specified precision.
func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return math.Round(num*output) / output
}
