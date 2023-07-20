package telemetry

import (
	"sync"

	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/models"
)

var (
	hardwareResources *HardwareResources
	mu                sync.Mutex
)

// HardwareResources is a struct with methods to simplify the hardware resources
// management. It has methods to change the resources and save on the DB.
type HardwareResources struct {
	DBFreeResources    models.FreeResources
	NewFreeRes         models.FreeResources
	OnboardedResources models.OnboardedResources
}

// IncreaseFreeResources calls modifyFreeResources to increase the FreeResources
// based on a models.Resource params
func (r *HardwareResources) IncreaseFreeResources(resourcesToModify models.Resources) {
	r.modifyFreeResources(resourcesToModify, 1)
}

// DecreaseFreeResources calls modifyFreeResources to decrease the FreeResources
// based on a models.Resource params
func (r *HardwareResources) DecreaseFreeResources(resourcesToModify models.Resources) {
	r.modifyFreeResources(resourcesToModify, -1)
}

// modifyFreeResources modifies the NewFreeRes struct, increasing/decreasing based on param received.
// This struct is the one which will be used to do a write operation on DB when calling the UpdateDBFreeResources.
func (r *HardwareResources) modifyFreeResources(resourcesToModify models.Resources, increaseOrDecrease int) {
	if resourcesToModify.TotCPU != 0 {
		r.NewFreeRes.TotCPU = r.NewFreeRes.TotCPU + resourcesToModify.TotCPU*models.MHz(increaseOrDecrease)
		// TODO: not sure if doing the right math for Vcpu here
		r.NewFreeRes.VCPU = r.NewFreeRes.TotCPU / r.OnboardedResources.CoreCPU
	}

	if resourcesToModify.RAM != 0 {
		r.NewFreeRes.RAM = r.NewFreeRes.RAM + resourcesToModify.RAM*models.MB(increaseOrDecrease)
	}

	if resourcesToModify.Disk != 0 {
		r.NewFreeRes.Disk = r.NewFreeRes.Disk + resourcesToModify.Disk*models.MB(increaseOrDecrease)
	}
}

// UpdateDBFreeResources updates/creates the FreeResources table
// based on HardwareResources.NewFreeRes struct
func (r *HardwareResources) UpdateDBFreeResources() error {
	// Check if we have a previous entry in the table
	r.NewFreeRes.ID = 1
	var freeRes models.FreeResources
	if res := db.DB.Find(&freeRes); res.RowsAffected == 0 {
		result := db.DB.Create(&r.NewFreeRes)
		if result.Error != nil {
			return result.Error
		}
	} else {
		result := db.DB.Model(&models.FreeResources{}).Where("id = ?", 1).Select("*").Updates(&r.NewFreeRes)
		if result.Error != nil {
			return result.Error
		}
	}
	r.DBFreeResources = r.NewFreeRes
	return nil
}

func NewHardwareResources() (*HardwareResources, error) {
	mu.Lock()
	defer mu.Unlock()

	if hardwareResources != nil {
		return hardwareResources, nil
	}

	var freeRes models.FreeResources
	freeRes, err := GetFreeResources()
	if err != nil {
		return nil, err
	}

	var onboardedRes models.OnboardedResources
	if res := db.DB.Find(&onboardedRes); res.RowsAffected == 0 {
		return nil, res.Error
	}

	hardwareResources = &HardwareResources{
		DBFreeResources:    freeRes,
		NewFreeRes:         freeRes,
		OnboardedResources: onboardedRes,
	}

	return hardwareResources, err
}

func GetFreeResources() (models.FreeResources, error) {
	var freeResource models.FreeResources
	if res := db.DB.Find(&freeResource); res.RowsAffected == 0 {
		return freeResource, res.Error
	}
	return freeResource, nil
}
