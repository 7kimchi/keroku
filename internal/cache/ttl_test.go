package cache

import (
	"testing"
	"time"
)

func TestTTLBoundary(t *testing.T) {
	c, clk := newTest(t, 10, time.Minute)
	c.Set("a", 1)
	clk.Advance(time.Minute - time.Nanosecond)
	if _, ok := c.Get("a"); !ok {
		t.Fatal("expired one nanosecond early")
	}
	clk.Advance(time.Nanosecond)
	if _, ok := c.Get("a"); ok {
		t.Fatal("still live at exact expiry")
	}
	if c.Len() != 0 {
		t.Fatalf("expired entry not removed on read, len %d", c.Len())
	}
}

func TestSetRefreshesTTL(t *testing.T) {
	c, clk := newTest(t, 10, time.Minute)
	c.Set("a", 1)
	clk.Advance(50 * time.Second)
	c.Set("a", 2)
	clk.Advance(50 * time.Second)
	if v, ok := c.Get("a"); !ok || v != 2 {
		t.Fatalf("refresh lost: %v %v", v, ok)
	}
}

func TestAddAfterExpiry(t *testing.T) {
	c, clk := newTest(t, 10, time.Second)
	if !c.Add("a", 1) || c.Add("a", 2) {
		t.Fatal("first Add must win and second must lose")
	}
	clk.Advance(time.Second)
	if !c.Add("a", 3) {
		t.Fatal("Add over expired entry lost")
	}
	if v, _ := c.Get("a"); v != 3 {
		t.Fatalf("got %d", v)
	}
}
