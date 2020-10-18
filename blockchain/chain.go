package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/dgraph-io/badger/v2"
)

type BlockChain struct {
	LastHash []byte // only used for initialization
	Database *badger.DB
}

const (
	genesisData = "First Transaction from Genesis"
	dbFile      = "./tmp/badger"
)

func DBexists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

func New(address string) *BlockChain {
	if DBexists() {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}
	var lastHash []byte

	db, err := badger.Open(badger.DefaultOptions(dbFile))
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			cb := CoinbaseTx(address, genesisData)
			genesis := Genesis(cb)
			fmt.Println("Genesis created")
			err = txn.Set(genesis.Hash, genesis.Serialize())
			Handle(err)
			err = txn.Set([]byte("lh"), genesis.Hash)

			lastHash = genesis.Hash

			return err
		}
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.ValueCopy(lastHash)
		return err
	})
	Handle(err)

	return &BlockChain{lastHash, db}
}

func Continue() *BlockChain {
	if DBexists() == false {
		fmt.Println("No existing blockchain found, create one!")
		runtime.Goexit()
	}

	var lastHash []byte

	db, err := badger.Open(badger.DefaultOptions(dbFile))
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.ValueCopy(lastHash)

		return err
	})
	Handle(err)

	return &BlockChain{lastHash, db}
}

func (c *BlockChain) AddBlock(txs []*Transaction) {
	var lastHash []byte

	err := c.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.ValueCopy(lastHash)

		return err
	})
	Handle(err)

	newBlock := NewBlock(lastHash, txs)

	err = c.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)

		c.LastHash = newBlock.Hash

		return err
	})
	Handle(err)
}

func (c *BlockChain) Iterator() *Iterator {
	return &Iterator{c.LastHash, c.Database}
}

func (c *BlockChain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unspentTxs []Transaction

	// 记录这个 address 在某次 tx output 得到的钱, 已经花掉
	// key: txID
	// value: outIndex
	STXOs := make(map[string]Set)

	iter := c.Iterator()
	for block := iter.Next(); block != nil; block = iter.Next() {
		for _, tx := range block.Transactions {
			// 进钱在先, 花钱在后, 所以不用担心这种顺序
			// todo: rename
			txID := hex.EncodeToString(tx.ID)

			for outIdx, out := range tx.Outputs {
				if STXOs[txID] != nil {
					// It appeared in a transaction's inputs before
					// indicating that it was spent
					if STXOs[txID].Has(outIdx) {
						continue
					}
				}
				if out.IsLockedWithKey(pubKeyHash) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}

			if tx.IsCoinbase() {
				continue
			}

			for _, in := range tx.Inputs {
				// Did you spend it? unlock
				// Umm.. Yes
				if in.IsUsedWithKey(pubKeyHash) {
					txID := hex.EncodeToString(in.TxID)
					STXOs[txID] = make(Set)
					STXOs[txID].Add(in.OutIndex)
				}
			}

		}

	}

	return unspentTxs
}

// UTXO: unspent transaction output
func (c *BlockChain) FindUTXO(pubKeyHash []byte) []TXOutput {
	// todo 重复搜寻了 outputs
	var UTXOs []TXOutput
	unspentTransactions := c.FindUnspentTransactions(pubKeyHash)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Outputs {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

func (c *BlockChain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	unspentTxs := c.FindUnspentTransactions(pubKeyHash)
	accumulated := 0

Work:
	// TODO refactor
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Outputs {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOuts[txID] = append(unspentOuts[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOuts
}

type Iterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

// From latest block to genesis
func (iter *Iterator) Next() *Block {
	if len(iter.CurrentHash) == 0 {
		return nil
	}

	var block *Block

	err := iter.Database.View(func(txn *badger.Txn) error {
		var encodedBlock []byte
		item, err := txn.Get(iter.CurrentHash)
		Handle(err)
		encodedBlock, err = item.ValueCopy(encodedBlock)
		block = Deserialize(encodedBlock)

		return err
	})
	Handle(err)

	iter.CurrentHash = block.PrevHash

	return block
}

func (c *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	iter := c.Iterator()
	for block := iter.Next(); block != nil; block = iter.Next() {
		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}
	}

	return Transaction{},
		errors.New("transaction does not exist")
}

func (c *BlockChain) SignTransaction(tx *Transaction, pk ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := c.FindTransaction(in.TxID)
		Handle(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(pk, prevTXs)
}

func (c *BlockChain) VerifyTransaction(tx *Transaction) bool {
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := c.FindTransaction(in.TxID)
		Handle(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}
