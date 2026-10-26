package main

import (
	"strconv"
	"testing"
)

// TestItoaLarge guards the crash where itoa overflowed its buffer for large
// print_error codes (recordHistory → errSignature → itoa), e.g. 0x08000000.
func TestItoaLarge(t *testing.T) {
	cases := []int{0, 1, 9, 10, 99, 100, 99999999, 100000000, 134217728, 2147483647, -1, -134217728}
	for _, n := range cases {
		if got, want := itoa(n), strconv.Itoa(n); got != want {
			t.Errorf("itoa(%d) = %q, want %q", n, got, want)
		}
	}
}
