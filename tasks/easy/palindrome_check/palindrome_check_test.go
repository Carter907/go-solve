package palindrome_check

import "testing"

func TestIsPalindrome(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"racecar", true},
		{"level", true},
		{"hello", false},
		{"", true},
		{"a", true},
		{"ab", false},
	}
	for _, tt := range tests {
		if got := IsPalindrome(tt.in); got != tt.want {
			t.Fatalf("IsPalindrome(%q) = %v; want %v", tt.in, got, tt.want)
		}
	}
}
