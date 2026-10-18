package bubbles

// Bubbles sorts a slice of integers using the bubble sort algorithm.
func Bubbles(data []int) {
	n := len(data)
	for i := 0; i < n; i++ {
		for j := 0; j < n-i-1; j++ {
			if data[j] > data[j+1] {
				data[j], data[j+1] = data[j+1], data[j]
			}
		}
	}
}
