package utils

// ToPointer returns a pointer to the given value
func ToPointer[T any](value T) *T {
	return &value
}
