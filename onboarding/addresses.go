package onboarding

import (
	"crypto/ecdsa"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fivebinaries/go-cardano-serialization/address"
	"github.com/fivebinaries/go-cardano-serialization/bip32"
	"github.com/fivebinaries/go-cardano-serialization/network"
	"github.com/tyler-smith/go-bip39"
	"gitlab.com/nunet/device-management-service/models"
)

// extra step needed because NewAddress panics on invalid address
func isValidCardano(addr string, valid* bool) {
	defer func() {
		if r := recover(); r != nil {
			*valid = false
		}
	}()
	if _, err := address.NewAddress(addr); err == nil {
		*valid = true
	}
}

func ValidateAddress(addr string) error {
	if common.IsHexAddress(addr) {
		return nil
	} else {
		var validCardano = false
		isValidCardano(addr, &validCardano)
		if validCardano {
			return nil
		} else {
			return errors.New("invalid address")
		}
	}
}

func GetEthereumAddressAndPrivateKey() (*models.BlockchainAddressPrivKey, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyString := hexutil.Encode(privateKeyBytes)
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("publicKey is not of type *ecdsa.PublicKey")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	var pair models.BlockchainAddressPrivKey

	pair.Address = address
	pair.PrivateKey = privateKeyString
	return &pair, nil
}

func harden(num uint) uint32 {
	return uint32(0x80000000 + num)
}

func GetCardanoAddressAndMnemonic() (*models.BlockchainAddressPrivKey, error) {
	var pair models.BlockchainAddressPrivKey
	entropy, _ := bip39.NewEntropy(256)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	pair.Mnemonic = mnemonic
	rootKey := bip32.FromBip39Entropy(
		entropy,
		[]byte{},
	)
	accountKey := rootKey.Derive(harden(1852)).Derive(harden(1815)).Derive(harden(0))
	utxoPubKey := accountKey.Derive(0).Derive(0).Public()
	utxoPubKeyHash := utxoPubKey.PublicKey().Hash()
	stakeKey := accountKey.Derive(2).Derive(0).Public()
	stakeKeyHash := stakeKey.PublicKey().Hash()
	addr := address.NewBaseAddress(
		network.MainNet(),
		&address.StakeCredential{
			Kind:    address.KeyStakeCredentialType,
			Payload: utxoPubKeyHash[:],
		},
		&address.StakeCredential{
			Kind:    address.KeyStakeCredentialType,
			Payload: stakeKeyHash[:],
		})
	pair.Address = addr.String()
	return &pair, nil
}
