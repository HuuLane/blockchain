package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
)

const (
	checksumLength = 4
	version        = byte(0x00)
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// e.g.
// Address: 1NPyKgykNjoZ6yTWfeazzY2vSHkAz9JoTx
// Full hash: 00eab2cedc386b1463f3a10b85b88a41b03f0c2bd5bc7eba1b
//
// checksumLength = 4 bytes = 8 char
// [Version]: 00
// [Pub key hash]: eab2cedc386b1463f3a10b85b88a41b03f0c2bd5
// [CheckSum]: bc7eba1b
func (w Wallet) Address() []byte {
	pubHash := PublicKeyHash(w.PublicKey)

	versionedHash := append([]byte{version}, pubHash...)
	checksum := Checksum(versionedHash)

	fullHash := append(versionedHash, checksum...)
	address := Base58Encode(fullHash)
	return address
}

func ValidateAddress(address string) bool {
	fullHash := Base58Decode([]byte(address))
	actualChecksum := fullHash[len(fullHash)-checksumLength:]
	version := fullHash[0]
	pubKeyHash := fullHash[1 : len(fullHash)-checksumLength]
	targetChecksum := Checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

func New() *Wallet {
	private, public := NewKeyPair()
	return &Wallet{private, public}
}

func Checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])

	return secondHash[:checksumLength]
}
