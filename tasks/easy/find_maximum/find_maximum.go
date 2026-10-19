package find_maximum

// FindMax returns the maximum integer in a non-empty slice of integers.
func FindMax(nums []int) int {
	if len(nums) == 0 {
		return 0
	}
	max := nums[0]
	for _, n := range nums[1:] {
		if n > max {
			max = n
		}
	}
	return max
}
