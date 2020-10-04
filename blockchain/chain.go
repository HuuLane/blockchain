package blockchain

import (
	"fmt"

	"github.com/dgraph-io/badger/v2"
)

type BlockChain struct {
	LastHash []byte // only used for initialization
	Database *badger.DB
}

var (
	singleton *BlockChain
)

func NewBlockChain() *BlockChain {
	if singleton == nil {
		var lastHash []byte

		db, err := badger.Open(badger.DefaultOptions("/tmp/badger"))
		Handle(err)

		err = db.Update(func(txn *badger.Txn) error {
			if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
				fmt.Println("No existing blockchain found")
				genesis := Genesis()
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

		singleton = &BlockChain{lastHash, db}
	}
	return singleton
}

func (c *BlockChain) AddBlock(data string) {
	var lastHash []byte

	err := c.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.ValueCopy(lastHash)

		return err
	})
	Handle(err)

	newBlock := NewBlock(lastHash, data)

	err = c.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)

		c.LastHash = newBlock.Hash

		return err
	})
	Handle(err)
}
