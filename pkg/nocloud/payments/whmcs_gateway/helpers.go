package whmcs_gateway

import (
	"math"
	"time"
)

const equalFloatsEpsilon = 1e-5

func ptr[T any](v T) *T {
	return &v
}

func equalFloats(a, b float64) bool {
	return math.Abs(a-b) < equalFloatsEpsilon
}

func compareFloat(a, b float64, precisionDigits int) bool {
	return math.Abs(a-b) < math.Pow10(-precisionDigits)
}

func isDateEqualWithoutTime(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}
