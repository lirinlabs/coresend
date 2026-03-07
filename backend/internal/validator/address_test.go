package validator

import (
	"math/rand"
	"strings"
	"testing"
)

func TestIsValidHexAddress(t *testing.T) {
	t.Parallel()

	valid := "0123456789abcdef0123456789abcdef01234567"

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "valid lowercase hex address",
			input: valid,
			want:  true,
		},
		{
			name:  "invalid empty string",
			input: "",
			want:  false,
		},
		{
			name:  "invalid length 39",
			input: strings.Repeat("a", 39),
			want:  false,
		},
		{
			name:  "invalid length 41",
			input: strings.Repeat("a", 41),
			want:  false,
		},
		{
			name:  "invalid uppercase letter",
			input: "0123456789Abcdef0123456789abcdef01234567",
			want:  false,
		},
		{
			name:  "invalid character g",
			input: "0123456789gbcdef0123456789abcdef01234567",
			want:  false,
		},
		{
			name:  "invalid punctuation",
			input: "0123456789!bcdef0123456789abcdef01234567",
			want:  false,
		},
		{
			name:  "invalid whitespace",
			input: "0123456789 bcdef0123456789abcdef01234567",
			want:  false,
		},
		{
			name:  "invalid newline",
			input: "0123456789\nbcdef0123456789abcdef01234567",
			want:  false,
		},
		{
			name:  "invalid 0x prefix",
			input: "0x0123456789abcdef0123456789abcdef01234567",
			want:  false,
		},
		{
			name:  "invalid mixed case",
			input: "0123456789abcDef0123456789abcdef01234567",
			want:  false,
		},
		{
			name:  "invalid all uppercase",
			input: strings.ToUpper(valid),
			want:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := IsValidHexAddress(tc.input)
			if got != tc.want {
				t.Fatalf("IsValidHexAddress(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestIsValidHexAddress_RandomInputsNoPanic(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewSource(42))

	for i := 0; i < 1000; i++ {
		n := rng.Intn(80)
		inputBytes := make([]byte, n)
		for j := range inputBytes {
			inputBytes[j] = byte(rng.Intn(256))
		}

		func(s string) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("IsValidHexAddress panicked for input %q: %v", s, r)
				}
			}()
			_ = IsValidHexAddress(s)
		}(string(inputBytes))
	}
}
