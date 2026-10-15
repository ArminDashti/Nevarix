package auth

import "testing"

func TestBearerMatches(t *testing.T) {
	cases := []struct {
		header, token string
		want          bool
	}{
		{"Bearer secret", "secret", true},
		{"Bearer wrong", "secret", false},
		{"secret", "secret", false},
		{"Bearer ", "", false},
		{"Bearer x", "", false},
	}
	for _, c := range cases {
		if got := BearerMatches(c.header, c.token); got != c.want {
			t.Errorf("BearerMatches(%q, %q) = %v, want %v", c.header, c.token, got, c.want)
		}
	}
}
