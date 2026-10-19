package sum_two

import "testing"

func TestSumTwo(t *testing.T) {
	tests := []struct {
		a, b int
		want int
	}{
		{1, 2, 3},
		{0, 0, 0},
		{-5, 5, 0},
		{100, -25, 75},
	}
	for _, tt := range tests {
		if got := SumTwo(tt.a, tt.b); got != tt.want {
			t.Fatalf("SumTwo(%d, %d) = %d; want %d", tt.a, tt.b, got, tt.want)
		}
	}
}
