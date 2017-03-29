package logger

import "testing"

func TestLoggerLogf(t *testing.T) {
	_, err := Logf("My message %s and for %s", "hello world", "John Smith")
	if err != nil {
		t.Fatal("Must not have error")
	}
}
