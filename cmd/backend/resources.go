package backend

import (
	"gitlab.com/nunet/device-management-service/dms/onboarding"
	"gitlab.com/nunet/device-management-service/models"
)

type Resources struct{}

func (r *Resources) GetTotalProvisioned() *models.Provisioned {
	return onboarding.GetTotalProvisioned()
}
