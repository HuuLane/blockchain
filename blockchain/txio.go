package blockchain

// An input of tx is an output of a previous tx
// So we need that tx id and that output index to trace the source of the money
// Pubkey is the address of the payee, spend money Need to use sig to sign
// but currently sig == pubkey
type TxInput struct {
	TxID     []byte // transaction ID
	OutIndex int    // index of output
	Sig      string //
}

type TxOutput struct {
	Value  int
	PubKey string
}

func (in *TxInput) CanUnlock(data string) bool {
	return in.Sig == data
}

func (out *TxOutput) CanBeUnlocked(data string) bool {
	return out.PubKey == data
}
