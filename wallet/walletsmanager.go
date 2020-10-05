package wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"io/ioutil"
	"log"
	"os"
)

const walletsFile = "./tmp/wallets.data"

type WalletsManager struct {
	Wallets map[string]*Wallet
}

func NewWalletsManager() (*WalletsManager, error) {
	wm := WalletsManager{}
	wm.Wallets = make(map[string]*Wallet)
	err := wm.Load()
	return &wm, err
}

func (wm *WalletsManager) AddWallet() string {
	wallet := New()
	address := string(wallet.Address())
	wm.Wallets[address] = wallet

	return address
}

func (wm *WalletsManager) GetAllAddresses() []string {
	var addresses []string

	for address := range wm.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

func (wm WalletsManager) GetWallet(address string) Wallet {
	return *wm.Wallets[address]
}

func (wm *WalletsManager) Load() error {
	if _, err := os.Stat(walletsFile); os.IsNotExist(err) {
		return err
	}

	var wallets WalletsManager

	fileContent, err := ioutil.ReadFile(walletsFile)
	if err != nil {
		return err
	}

	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		return err
	}

	wm.Wallets = wallets.Wallets

	return nil
}

func (wm *WalletsManager) Save() {
	var content bytes.Buffer

	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(wm)
	if err != nil {
		log.Panic(err)
	}

	err = ioutil.WriteFile(walletsFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
