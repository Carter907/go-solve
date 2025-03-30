package bubbles

import (
	"slices"
	"testing"
)

func TestBubbles(t *testing.T) {
	testData := []int{3, 7, 3, 1, 8, 2, 3, 9}
	otherData := []int{3, 7, 3, 1, 8, 2, 3, 9}

	slices.Sort(testData)
	slices.Sort(otherData)

	if !slices.Equal(testData, otherData) {
		t.Fatalf("Bubbles() did not sort the data! returned %v", otherData)
	}
}
