package phone

import "testing"

func TestCacheKeyIsVersioned(t *testing.T) {
	got := cacheKey("+84705899899")
	if got != "phone:v1:+84705899899" {
		t.Fatalf("unexpected cache key: %s", got)
	}
}
