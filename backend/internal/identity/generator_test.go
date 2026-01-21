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

func TestAddressFromMnemonic(t *testing.T) {
	tests := []struct {
		name     string
		mnemonic string
		expected string
	}{
		{
			name:     "standard mnemonic",
			mnemonic: "witch collapse practice feed shame open despair creek road again ice least",
			expected: "b4ebe3e2200cbc90",
		},
		{
			name:     "lowercase input",
			mnemonic: "witch collapse practice feed shame open despair creek road again ice least",
			expected: "b4ebe3e2200cbc90",
		},
		{
			name:     "uppercase input",
			mnemonic: "WITCH COLLAPSE PRACTICE FEED SHAME OPEN DESPAIR CREEK ROAD AGAIN ICE LEAST",
			expected: "b4ebe3e2200cbc90",
		},
		{
			name:     "mixed case input",
			mnemonic: "Witch Collapse Practice Feed Shame Open Despair Creek Road Again Ice Least",
			expected: "b4ebe3e2200cbc90",
		},
		{
			name:     "with extra spaces",
			mnemonic: "  witch collapse practice feed shame open despair creek road again ice least  ",
			expected: "b4ebe3e2200cbc90",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AddressFromMnemonic(tt.mnemonic)
			if result != tt.expected {
				t.Errorf("AddressFromMnemonic() = %v, want %v", result, tt.expected)
			}

			if len(result) != 16 {
				t.Errorf("AddressFromMnemonic() returned address of length %d, expected 16", len(result))
			}
		})
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

	expected := "e3b0c44298fc1c14"
	if result != expected {
		t.Errorf("AddressFromMnemonic() empty string = %v, want %v", result, expected)
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
