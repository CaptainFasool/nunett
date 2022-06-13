package models

// AddressPrivKey holds Ethereum wallet address and private key from which the
// address is derived.
type AddressPrivKey struct {
	Address    string
	PrivateKey string
}
