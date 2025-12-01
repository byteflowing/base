package trans

func Deref[T any](ptr *T) T {
	if ptr == nil {
		var zero T
		return zero
	}
	return *ptr
}

func Ref[T any](v T) *T {
	return &v
}
