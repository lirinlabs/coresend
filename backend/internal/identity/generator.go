package identity

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"

	"github.com/tyler-smith/go-bip39"
)

const AddressLength = 16

var validAddressRegex = regexp.MustCompile(`^[a-f0-9]{16}$`)

func GenerateNewMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return "", err
	}
	return bip39.NewMnemonic(entropy)
}

func AddressFromMnemonic(mnemonic string) string {
	mnemonic = strings.TrimSpace(strings.ToLower(mnemonic))
	hash := sha256.Sum256([]byte(mnemonic))
	return hex.EncodeToString(hash[:])[:AddressLength]
}

func IsValidAddress(addr string) bool {
	return validAddressRegex.MatchString(strings.ToLower(addr))
}

func IsValidBIP39Mnemonic(mnemonic string) bool {
	mnemonic = strings.TrimSpace(strings.ToLower(mnemonic))
	_, err := bip39.EntropyFromMnemonic(mnemonic)
	return err == nil
}
