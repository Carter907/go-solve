package hello_world

import "testing"

func TestHelloWorld(t *testing.T) {
	if message := ReturnMessage(); message != "Hello World" {
		t.Fatalf("ReturnMessage() = %q, want \"Hello World\"", message)
	}
}
