package whmcs_gateway

func ptr[T any](v T) *T {
	return &v
}
