package identity

import (
	"strings"
	"testing"
)

func TestGenerateNewMnemonic(t *testing.T) {
	mnemonic, err := GenerateNewMnemonic()
	if err != nil {
		t.Fatalf("GenerateNewMnemonic() error = %v", err)
	}

	if mnemonic == "" {
		t.Fatal("GenerateNewMnemonic() returned empty string")
	}

	words := strings.Split(mnemonic, " ")
	if len(words) != 12 {
		t.Fatalf("Expected 12 words, got %d", len(words))
	}

	mnemonic2, err := GenerateNewMnemonic()
	if err != nil {
		t.Fatalf("GenerateNewMnemonic() second call error = %v", err)
	}

	if mnemonic == mnemonic2 {
		t.Fatal("GenerateNewMnemonic() returned same mnemonic twice")
	}
}

func TestAddressFromMnemonic_Deterministic(t *testing.T) {
	mnemonic := "witch collapse practice feed shame open despair creek road again ice least"

	result1 := AddressFromMnemonic(mnemonic)
	result2 := AddressFromMnemonic(mnemonic)

	if result1 != result2 {
		t.Errorf("AddressFromMnemonic() not deterministic: %s != %s", result1, result2)
	}
}

func TestAddressFromMnemonic_EmptyString(t *testing.T) {
	result := AddressFromMnemonic("")

	if len(result) != 16 {
		t.Errorf("AddressFromMnemonic() empty string returned length %d, want 16", len(result))
	}
}

func TestIsValidAddress(t *testing.T) {
	tests := []struct {
		name     string
		addr     string
		expected bool
	}{
		{"valid lowercase", "b4ebe3e2200cbc90", true},
		{"valid uppercase", "B4EBE3E2200CBC90", true},
		{"valid mixed case", "B4ebe3E2200cbc90", true},
		{"too short", "b4ebe3e2200cbc9", false},
		{"too long", "b4ebe3e2200cbc901", false},
		{"invalid chars", "b4ebe3e2200cbc9g", false},
		{"empty string", "", false},
		{"with spaces", "b4ebe3e2 00cbc90", false},
		{"valid from mnemonic", AddressFromMnemonic("witch collapse practice feed shame open despair creek road again ice least"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidAddress(tt.addr)
			if result != tt.expected {
				t.Errorf("IsValidAddress(%q) = %v, want %v", tt.addr, result, tt.expected)
			}
		})
	}
}
