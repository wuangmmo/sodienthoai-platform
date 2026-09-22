package phone

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct{ in, want string; ok bool }{
		{"+84705899899", "+84705899899", true},
		{"+84 705 899 899", "+84705899899", true},
		{"+84 (705) 899-899", "+84705899899", true},
		{"0705899899", "", false},
		{"+0123456789", "", false},
		{"+84abc", "", false},
	}
	for _, tt := range tests {
		got, err := Normalize(tt.in)
		if tt.ok && (err != nil || got != tt.want) { t.Fatalf("%q: got %q err=%v",tt.in,got,err) }
		if !tt.ok && err == nil { t.Fatalf("%q: expected error, got %q",tt.in,got) }
	}
}
