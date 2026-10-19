package word_count

import (
	"reflect"
	"testing"
)

func TestWordCount(t *testing.T) {
	got := WordCount("go is fun and go is fast")
	want := map[string]int{
		"go":   2,
		"is":   2,
		"fun":  1,
		"and":  1,
		"fast": 1,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("WordCount() = %#v; want %#v", got, want)
	}

	gotEmpty := WordCount("")
	if len(gotEmpty) != 0 {
		t.Fatalf("WordCount(\"\") = %#v; want empty map", gotEmpty)
	}
}
