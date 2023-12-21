package backend

import (
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/onboarding"
)

type Resources struct{}

func (r *Resources) GetTotalProvisioned() *models.Provisioned {
	return onboarding.GetTotalProvisioned()
}
