package cache

import (
	"testing"
	"time"
)

func FuzzCacheOps(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5})
	f.Fuzz(func(t *testing.T, ops []byte) {
		c, clk := newTest(t, 8, time.Second)
		shadow := map[string]int{}
		for i, op := range ops {
			k := string(rune('a' + op%12))
			switch op % 5 {
			case 0:
				c.Set(k, i)
				shadow[k] = i
			case 1:
				if v, ok := c.Get(k); ok && v != shadow[k] {
					t.Fatalf("key %q got %d want %d", k, v, shadow[k])
				}
			case 2:
				c.Delete(k)
				delete(shadow, k)
			case 3:
				clk.Advance(300 * time.Millisecond)
			default:
				if c.Add(k, i) {
					shadow[k] = i
				}
			}
			if c.Len() > 8 {
				t.Fatalf("len %d over bound", c.Len())
			}
		}
	})
}
