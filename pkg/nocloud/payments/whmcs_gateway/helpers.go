package whmcs_gateway

import "math"

const equalFloatsEpsilon = 1e-5

func ptr[T any](v T) *T {
	return &v
}

func equalFloats(a, b float64) bool {
	return math.Abs(a-b) < equalFloatsEpsilon
}
