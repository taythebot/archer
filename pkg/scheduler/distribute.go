package scheduler

// roundRobin equally distributes an array of targets to N workers
func roundRobin(targets []string, workers int) (result [][]string) {
	// Determine number of chunks per pass: ceil(len(list) / n)
	chunkSize := (len(targets) + workers - 1) / workers

	// Distribute
	for i := 0; i < len(targets); i += chunkSize {
		// Get last index
		end := i + chunkSize

		// If at end, set last index to array length
		if end > len(targets) {
			end = len(targets)
		}

		// Append to results
		result = append(result, targets[i:end])
	}

	return
}
