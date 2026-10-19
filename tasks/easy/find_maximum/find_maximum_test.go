package find_maximum

import "testing"

func TestFindMax(t *testing.T) {
	tests := []struct {
		in   []int
		want int
	}{
		{[]int{1, 5, 3, 9, 2}, 9},
		{[]int{-10, -3, -50, -1}, -1},
		{[]int{42}, 42},
		{[]int{7, 7, 7}, 7},
	}
	for _, tt := range tests {
		if got := FindMax(tt.in); got != tt.want {
			t.Fatalf("FindMax(%v) = %d; want %d", tt.in, got, tt.want)
		}
	}
}
