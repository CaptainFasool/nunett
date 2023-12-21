package backend

import (
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/onboarding"
)

type Wallet struct{}

func (w *Wallet) GetCardanoAddressAndMnemonic() (*models.BlockchainAddressPrivKey, error) {
	return onboarding.GetCardanoAddressAndMnemonic()
}

func (w *Wallet) GetEthereumAddressAndPrivateKey() (*models.BlockchainAddressPrivKey, error) {
	return onboarding.GetEthereumAddressAndPrivateKey()
}
