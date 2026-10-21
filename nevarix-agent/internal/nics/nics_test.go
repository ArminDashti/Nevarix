package nics

import "testing"

func TestRate(t *testing.T) {
	if got := Rate(2000, 1000, 2); got != 500 {
		t.Fatalf("Rate = %v want 500", got)
	}
	if got := Rate(100, 200, 1); got != 0 {
		t.Fatalf("negative delta Rate = %v want 0", got)
	}
}
