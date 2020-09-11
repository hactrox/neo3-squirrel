package models

// AddressInfo db model.
type AddressInfo struct {
	ID          uint
	Address     string
	FirstTxTime uint64
	LastTxTime  uint64
	Transfers   uint
}
