package blockchain

import (
	"encoding/hex"
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

func (c *BlockChain) FindUnspentTransactions(address string) []Transaction {
	var unspentTxs []Transaction

	// todo: refactor to set
	// 记录这个 address 在某次 tx output 得到的钱, 已经花掉
	// key: txID
	// value: outIndex
	STXOs := make(map[string]map[int]struct{})

	iter := c.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			// 进钱在先, 花钱在后, 所以不用担心这种顺序
			// todo: rename
			txID := hex.EncodeToString(tx.ID)

			for outIdx, out := range tx.Outputs {
				if STXOs[txID] != nil {
					// It appeared in a transaction's inputs before
					// indicating that it was spent
					if _, ok := STXOs[txID][outIdx]; ok {
						continue
					}
				}
				if out.CanBeUnlocked(address) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}

			if tx.IsCoinbase() {
				continue
			}

			for _, in := range tx.Inputs {
				// Did you spend it? unlock
				// Umm.. Yes
				if in.CanUnlock(address) {
					txID := hex.EncodeToString(in.TxID)
					STXOs[txID] = make(map[int]struct{})
					STXOs[txID][in.OutIndex] = struct{}{}
				}
			}

		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return unspentTxs
}

// UTXO: unspent transaction output
func (c *BlockChain) FindUTXO(address string) []TxOutput {
	var UTXOs []TxOutput
	unspentTransactions := c.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Outputs {
			if out.CanBeUnlocked(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

func (c *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	unspentTxs := c.FindUnspentTransactions(address)
	accumulated := 0

Work:
	// TODO refactor
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Outputs {
			if out.CanBeUnlocked(address) && accumulated < amount {
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

// todo: using nil to terminal
func (iter *Iterator) Next() *Block {
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
