package reverse_string

import "testing"

func TestReverseString(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"hello", "olleh"},
		{"Go", "oG"},
		{"", ""},
		{"racecar", "racecar"},
		{"golang", "gnalog"},
	}
	for _, tt := range tests {
		if got := ReverseString(tt.in); got != tt.want {
			t.Fatalf("ReverseString(%q) = %q; want %q", tt.in, got, tt.want)
		}
	}
}
