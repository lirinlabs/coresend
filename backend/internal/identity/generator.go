package identity

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"

	"github.com/tyler-smith/go-bip39"
)

const AddressLength = 16

const (
	DomainSeparationKey = "coresend-auth"
	SignatureSeparator  = "|"
)

var validAddressRegex = regexp.MustCompile(`^[a-f0-9]{16}$`)

func GenerateNewMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return "", err
	}
	return bip39.NewMnemonic(entropy)
}

func AddressFromMnemonic(mnemonic string) string {
	_, pubkey, err := DeriveEd25519KeyPair(mnemonic)
	if err != nil {
		return ""
	}
	return AddressFromPublicKey(pubkey)
}

func IsValidAddress(addr string) bool {
	return validAddressRegex.MatchString(strings.ToLower(addr))
}

func IsValidBIP39Mnemonic(mnemonic string) bool {
	mnemonic = strings.TrimSpace(strings.ToLower(mnemonic))
	_, err := bip39.EntropyFromMnemonic(mnemonic)
	return err == nil
}

func DeriveEd25519KeyPair(mnemonic string) ([]byte, []byte, error) {
	mnemonic = strings.TrimSpace(strings.ToLower(mnemonic))

	h := hmac.New(sha256.New, []byte(DomainSeparationKey))
	h.Write([]byte(mnemonic))
	seed := h.Sum(nil)

	if len(seed) < 32 {
		seed = seed[:32]
	}

	privkey := ed25519.NewKeyFromSeed(seed)
	pubkey := privkey.Public().(ed25519.PublicKey)

	return privkey, pubkey, nil
}

func AddressFromPublicKey(pubkey []byte) string {
	hash := sha256.Sum256(pubkey)
	return hex.EncodeToString(hash[:])[:AddressLength]
}

func VerifySignature(pubkey []byte, message string, signature []byte) bool {
	return ed25519.Verify(pubkey, []byte(message), signature)
}

func CreateMessageToSign(address string, timestamp int64) string {
	return strings.ToLower(address) + SignatureSeparator + string(rune(timestamp))
}
