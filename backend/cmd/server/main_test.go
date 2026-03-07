package main

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		fallback string
		setup    func(t *testing.T, key string)
		want     string
	}{
		{
			name:     "returns env value when set",
			key:      "CORESEND_TEST_GETENV_SET",
			fallback: "fallback",
			setup: func(t *testing.T, key string) {
				t.Setenv(key, "configured")
			},
			want: "configured",
		},
		{
			name:     "returns fallback when var unset",
			key:      "CORESEND_TEST_GETENV_UNSET",
			fallback: "fallback",
			setup: func(t *testing.T, key string) {
				unsetEnvForTest(t, key)
			},
			want: "fallback",
		},
		{
			name:     "returns fallback when var empty",
			key:      "CORESEND_TEST_GETENV_EMPTY",
			fallback: "fallback",
			setup: func(t *testing.T, key string) {
				t.Setenv(key, "")
			},
			want: "fallback",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(t, tc.key)

			got := getEnv(tc.key, tc.fallback)
			if got != tc.want {
				t.Fatalf("getEnv(%q, %q) = %q, want %q", tc.key, tc.fallback, got, tc.want)
			}
		})
	}
}

func unsetEnvForTest(t *testing.T, key string) {
	t.Helper()

	originalValue, hadValue := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("failed to unset %q: %v", key, err)
	}

	t.Cleanup(func() {
		if hadValue {
			if err := os.Setenv(key, originalValue); err != nil {
				t.Fatalf("failed to restore %q: %v", key, err)
			}
		}
	})
}
