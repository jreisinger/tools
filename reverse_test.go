package tools

import "testing"

func TestReverse(t *testing.T) {
	testcases := []struct {
		s    string
		want string
	}{
		{"", ""},
		{"a", "a"},
		{"123", "321"},
		{"abba", "abba"},
		{"Hello 世界", "界世 olleH"},
	}
	for _, tc := range testcases {
		for _, f := range []func(s string) string{Reverse, Reverse2} {
			test(t, f, tc.s, tc.want)
		}
	}
}

func test(t *testing.T, reverse func(s string) string, s, want string) {
	got := reverse(s)
	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}
