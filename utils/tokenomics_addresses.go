package utils

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/fivebinaries/go-cardano-serialization/address"
)

// extra step needed because NewAddress panics on invalid address
func isValidCardano(addr string, valid *bool) {
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
		return errors.New("ethereum wallet address not allowed")
	}

	var validCardano = false
	isValidCardano(addr, &validCardano)
	if validCardano {
		return nil
	} else {
		return errors.New("invalid cardano wallet address")
	}
}
