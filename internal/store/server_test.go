package store

import (
	"strings"
	"testing"
)

func TestRequireServer(t *testing.T) {
	for _, n := range []int{180000, 180006, 190000, 1000000} {
		if err := requireServer(n); err != nil {
			t.Fatalf("%d refused: %v", n, err)
		}
	}
	for n, want := range map[int]string{170009: "17.9", 160015: "16.15", 140000: "14.0", 0: "0.0", -1: "0.-1"} {
		err := requireServer(n)
		if err == nil || !strings.Contains(err.Error(), "server is "+want) {
			t.Fatalf("%d: %v", n, err)
		}
	}
}
