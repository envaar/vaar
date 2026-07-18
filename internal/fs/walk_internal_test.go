package fs

import "testing"

func TestIsDotenvFile(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{name: ".env", want: true},
		{name: ".env.local", want: true},
		{name: ".env.prod", want: true},
		{name: ".env.temp", want: true},
		{name: ".env.preview-local", want: true},
		{name: ".env.", want: false},
		{name: ".environment", want: false},
		{name: "my.env", want: false},
		{name: ".envrc", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isDotenvFile(tc.name); got != tc.want {
				t.Fatalf("isDotenvFile(%q) = %t, want %t", tc.name, got, tc.want)
			}
		})
	}
}
