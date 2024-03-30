package backend

import (
	"gitlab.com/nunet/device-management-service/dms/onboarding"
	"gitlab.com/nunet/device-management-service/models"
)

type Wallet struct{}

func (w *Wallet) GetCardanoAddressAndMnemonic() (*models.BlockchainAddressPrivKey, error) {
	return onboarding.GetCardanoAddressAndMnemonic()
}

func (w *Wallet) GetEthereumAddressAndPrivateKey() (*models.BlockchainAddressPrivKey, error) {
	return onboarding.GetEthereumAddressAndPrivateKey()
}
