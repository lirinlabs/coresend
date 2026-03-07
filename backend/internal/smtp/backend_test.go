package smtp

import "testing"

func TestExtractLocalPart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "standard email",
			input: "alice@example.com",
			want:  "alice",
		},
		{
			name:  "no at symbol",
			input: "no-at-symbol",
			want:  "no-at-symbol",
		},
		{
			name:  "multiple at symbols uses last index",
			input: "x@y@z",
			want:  "x@y",
		},
		{
			name:  "empty local part",
			input: "@example.com",
			want:  "",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := extractLocalPart(tc.input)
			if got != tc.want {
				t.Fatalf("extractLocalPart(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
