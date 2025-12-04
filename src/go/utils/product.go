package utils

func CartesianProduct[T any](lists [][]T) [][]T {
	if len(lists) == 0 {
		return [][]T{}
	}

	// Start with a single empty combination as the initial result
	result := [][]T{{}}
	for _, currentList := range lists {
		var nextResult [][]T
		for _, combo := range result {
			for _, item := range currentList {
				// Create a new combination by appending the current item
				// to the existing combination
				newCombo := make([]T, len(combo))
				copy(newCombo, combo)
				newCombo = append(newCombo, item)
				nextResult = append(nextResult, newCombo)
			}
		}
		result = nextResult
	}

	return result
}
