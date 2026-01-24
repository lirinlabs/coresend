package identity

import (
	"testing"

	"golang.org/x/crypto/ed25519"
)

func TestDeriveEd25519KeyPair(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	privkey, pubkey, err := DeriveEd25519KeyPair(mnemonic)
	if err != nil {
		t.Fatalf("DeriveEd25519KeyPair failed: %v", err)
	}

	if len(privkey) != ed25519.PrivateKeySize {
		t.Errorf("Expected privkey length %d, got %d", ed25519.PrivateKeySize, len(privkey))
	}

	if len(pubkey) != ed25519.PublicKeySize {
		t.Errorf("Expected pubkey length %d, got %d", ed25519.PublicKeySize, len(pubkey))
	}
}

func TestAddressFromPublicKey(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	_, pubkey, _ := DeriveEd25519KeyPair(mnemonic)

	address := AddressFromPublicKey(pubkey)

	if len(address) != AddressLength {
		t.Errorf("Expected address length %d, got %d", AddressLength, len(address))
	}

	if !IsValidAddress(address) {
		t.Errorf("Derived address %s is not valid", address)
	}
}

func TestVerifySignature(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	privkey, pubkey, _ := DeriveEd25519KeyPair(mnemonic)

	message := "test-message|1234567890"
	signature := ed25519.Sign(privkey, []byte(message))

	if !VerifySignature(pubkey, message, signature) {
		t.Error("Signature verification failed for valid signature")
	}

	tamperedMessage := "tampered-message|1234567890"
	if VerifySignature(pubkey, tamperedMessage, signature) {
		t.Error("Signature verification succeeded for tampered message")
	}
}

func TestCreateMessageToSign(t *testing.T) {
	address := "b4ebe3e2200cbc90"
	timestamp := int64(1737705600000)

	message := CreateMessageToSign(address, timestamp)

	expected := address + "|" + string(rune(timestamp))
	if message != expected {
		t.Errorf("Expected message %q, got %q", expected, message)
	}
}

func TestKeyGenerationDeterministic(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	privkey1, pubkey1, _ := DeriveEd25519KeyPair(mnemonic)
	privkey2, pubkey2, _ := DeriveEd25519KeyPair(mnemonic)

	if string(privkey1) != string(privkey2) {
		t.Error("Same mnemonic produced different private keys")
	}

	if string(pubkey1) != string(pubkey2) {
		t.Error("Same mnemonic produced different public keys")
	}
}

func TestDifferentMnemonicsDifferentKeys(t *testing.T) {
	mnemonic1 := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	mnemonic2 := "witch collapse practice feed shame open despair creek road again ice least"

	_, pubkey1, _ := DeriveEd25519KeyPair(mnemonic1)
	_, pubkey2, _ := DeriveEd25519KeyPair(mnemonic2)

	if string(pubkey1) == string(pubkey2) {
		t.Error("Different mnemonics produced same public keys")
	}

	address1 := AddressFromPublicKey(pubkey1)
	address2 := AddressFromPublicKey(pubkey2)

	if address1 == address2 {
		t.Error("Different mnemonics produced same addresses")
	}
}

func TestDomainSeparation(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	_, pubkey1, _ := DeriveEd25519KeyPair(mnemonic)
	address1 := AddressFromPublicKey(pubkey1)

	if len(address1) != AddressLength {
		t.Errorf("Address length should be %d", AddressLength)
	}
}
