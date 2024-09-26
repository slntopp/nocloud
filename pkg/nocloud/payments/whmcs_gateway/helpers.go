package whmcs_gateway

func ptr[T any](v T) *T {
	if v == nil {
		return nil
	}
	return &v
}
