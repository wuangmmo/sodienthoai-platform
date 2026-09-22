package phone

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		in, want string
		ok       bool
	}{
		{"+84705899899", "+84705899899", true},
		{"+84 705 899 899", "+84705899899", true},
		{"+84 (705) 899-899", "+84705899899", true},
		{"0705899899", "", false},
		{"+0123456789", "", false},
		{"+84abc", "", false},
		{"＋84705899899", "", false},
		{"+８４７０５８９９８９９", "", false},
	}
	for _, tt := range tests {
		got, err := Normalize(tt.in)
		if tt.ok && (err != nil || got != tt.want) {
			t.Fatalf("%q: got %q err=%v", tt.in, got, err)
		}
		if !tt.ok && err == nil {
			t.Fatalf("%q: expected error, got %q", tt.in, got)
		}
	}
}

func TestNormalizeForVietnam(t *testing.T) {
	for _, in := range []string{"0705899899", "0705 899 899", "0705-899-899"} {
		got, err := NormalizeForCountry(in, "VN")
		if err != nil || got != "+84705899899" {
			t.Fatalf("%q: got %q err=%v", in, got, err)
		}
	}
	for _, in := range []string{"00705899899", "000705899899"} {
		if _, err := NormalizeForCountry(in, "VN"); err == nil {
			t.Fatalf("%q: repeated trunk prefix must be rejected", in)
		}
	}
	if _, err := NormalizeForCountry("0705899899", "US"); err == nil {
		t.Fatal("national number must not be guessed outside explicit VN context")
	}
}
