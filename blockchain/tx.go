package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/HuuLane/stupidcoin/wallet"
)

type Transaction struct {
	ID      []byte // checksum as ID
	Inputs  []TXInput
	Outputs []TXOutput
}

func NewTransaction(from, to string, amount int, chain *BlockChain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	wm, err := wallet.NewWalletsManager()
	Handle(err)
	w := wm.GetWallet(from)
	pubKeyHash := wallet.PublicKeyHash(w.PublicKey)
	acc, validOutputs := chain.FindSpendableOutputs(pubKeyHash, amount)

	if acc < amount {
		log.Panic("Error: not enough funds")
	}

	for txId, outs := range validOutputs {
		txID, err := hex.DecodeString(txId)
		Handle(err)

		for _, out := range outs {
			input := TXInput{txID, out, nil, w.PublicKey}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, *NewTXOutput(amount, to))

	// Return back the remaining money
	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from))
	}

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	chain.SignTransaction(&tx, w.PrivateKey)

	return &tx
}

func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

func CoinbaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Coins to %s", to)
	}

	txIn := TXInput{[]byte{}, -1, nil, []byte(data)}
	txOut := NewTXOutput(100, to)

	tx := Transaction{nil, []TXInput{txIn}, []TXOutput{*txOut}}
	tx.ID = tx.Hash()

	return &tx
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].TxID) == 0 && tx.Inputs[0].OutIndex == -1
}

func (tx *Transaction) Sign(pk ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}
	// Whether the sign can succeeds depends on whether the input is legal
	for _, in := range tx.Inputs {
		if _, ok := prevTXs[hex.EncodeToString(in.TxID)]; !ok {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	for inIdx, in := range txCopy.Inputs {
		prevTX := prevTXs[hex.EncodeToString(in.TxID)]
		txCopy.Inputs[inIdx].Signature = nil
		txCopy.Inputs[inIdx].PubKey = prevTX.Outputs[in.OutIndex].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inIdx].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &pk, txCopy.ID)
		Handle(err)
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Inputs[inIdx].Signature = signature
	}
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	for _, in := range tx.Inputs {
		if _, ok := prevTXs[hex.EncodeToString(in.TxID)]; !ok {
			log.Panic("Previous transaction not exist")
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inId, in := range tx.Inputs {
		prevTx := prevTXs[hex.EncodeToString(in.TxID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PubKey = prevTx.Outputs[in.OutIndex].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inId].PubKey = nil

		r := big.Int{}
		s := big.Int{}

		sigLen := len(in.Signature)
		r.SetBytes(in.Signature[:(sigLen / 2)])
		s.SetBytes(in.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(in.PubKey)
		x.SetBytes(in.PubKey[:(keyLen / 2)])
		y.SetBytes(in.PubKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
			return false
		}
	}

	return true
}

func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, in := range tx.Inputs {
		inputs = append(inputs, TXInput{in.TxID, in.OutIndex, nil, nil})
	}

	for _, out := range tx.Outputs {
		outputs = append(outputs, TXOutput{out.Value, out.PubKeyHash})
	}

	return Transaction{tx.ID, inputs, outputs}
}

func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))
	for i, input := range tx.Inputs {
		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:     %x", input.TxID))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.OutIndex))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.Outputs {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}
