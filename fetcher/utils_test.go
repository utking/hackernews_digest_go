package fetcher

import "testing"

func TestUtilsMin(t *testing.T) {
	if min(1, 10) != 1 {
		t.Fatal("min(1, 10) must be 1")
	}
	if min(1, 0) != 0 {
		t.Fatal("min(1, 0) must be 0")
	}
}
