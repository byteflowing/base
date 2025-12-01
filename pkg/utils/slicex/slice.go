package slicex

func Unique[T comparable](slice []T) []T {
	uniqueMap := make(map[T]struct{})
	var result []T
	for _, item := range slice {
		if _, exists := uniqueMap[item]; !exists {
			uniqueMap[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
