package dms

import (
	"context"
	"fmt"

	"gitlab.com/nunet/device-management-service/dms/resources"
	"gitlab.com/nunet/device-management-service/docker"
	"gitlab.com/nunet/device-management-service/models"
	"gorm.io/gorm"
)

// SanityCheck does basic consistency checks before starting the DMS in the following sequence:
// It checks for services that are marked running from the database and stops then removes them.
// Update their status to 'finshed with errors'.
// Recalculates free resources and update the database.
func SanityCheck(gormDB *gorm.DB) {
	var services []models.Services
	result := gormDB.Where("job_status = ?", "running").Find(&services)
	if result.Error != nil {
		panic(fmt.Errorf("unable to query running containers - %v", result.Error))
	}

	for _, service := range services {
		zlog.Sugar().Infof("Killing container %s", service.ContainerID)
		err := docker.StopAndRemoveContainer(context.Background(), service.ContainerID)
		if err != nil {
			zlog.Sugar().Errorf("Unable to kill container %s: %s", service.ContainerID, err)
			continue
		}
		service.JobStatus = "finished with errors"
		gormDB.Save(&service)
	}

	resources.CalcFreeResAndUpdateDB()
}
