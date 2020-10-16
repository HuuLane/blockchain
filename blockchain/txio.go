package blockchain

import (
	"bytes"

	"github.com/HuuLane/stupidcoin/wallet"
)

// An input of tx is an output of a previous tx
// So we need that tx id and that output index to trace the source of the money
type TXInput struct {
	TxID      []byte // transaction ID
	OutIndex  int    // index of output
	Signature []byte
	PubKey    []byte
}

func (in *TXInput) IsUsedWithKey(pubKeyHash []byte) bool {
	lockingHash := wallet.PublicKeyHash(in.PubKey)
	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

// what is PubKeyHash?
// In BTC, the recipient's receiving address is the hash of his public key
// The public key will be shown when he wants to use the money.
// It's easy to verify.
type TXOutput struct {
	Value      int
	PubKeyHash []byte
}

func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock(address)

	return txo
}

func (out *TXOutput) Lock(address string) {
	_, pubKeyHash, _ := wallet.ParseAddress(address)
	out.PubKeyHash = pubKeyHash
}

func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}
