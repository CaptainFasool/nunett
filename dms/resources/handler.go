package resources

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/models"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

// CalcFreeResources godoc
//
//	@Summary		Returns the amount of free resources available
//	@Description	Checks and returns the amount of free resources available
//	@Tags			telemetry
//	@Produce		json
//	@Success		200
//	@Router			/telemetry/free [get]
func GetFreeResource(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("URL", "/telemetry/free"))

	err := CalcFreeResAndUpdateDB()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	var freeResource models.FreeResources
	if res := db.DB.WithContext(c.Request.Context()).Find(&freeResource); res.RowsAffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error})
		return
	}

	c.JSON(http.StatusOK, freeResource)

}

func updateDBFreeResources(freeRes models.FreeResources) error {
	freeRes.ID = 1 // Enforce unique record for a given machine

	var freeResourcesModel models.FreeResources
	if res := db.DB.Find(&freeResourcesModel); res.RowsAffected == 0 {
		result := db.DB.Create(&freeRes)
		if result.Error != nil {
			return result.Error
		}
	} else {
		result := db.DB.Model(&models.FreeResources{}).Where("id = ?", 1).Select("*").Updates(&freeRes)
		if result.Error != nil {
			return result.Error
		}
	}
	return nil
}

func getServiceResourcesRequirements(gormDB *gorm.DB) (map[int]models.ServiceResourceRequirements, error) {
	var serviceResRequirements []models.ServiceResourceRequirements
	result := gormDB.Find(&serviceResRequirements)
	if result.Error != nil {
		return nil, fmt.Errorf("unable to query resource requirements - %v", result.Error)
	}

	mappedServicesResRequirements := make(map[int]models.ServiceResourceRequirements)
	for _, rr := range serviceResRequirements {
		mappedServicesResRequirements[int(rr.ID)] = rr
	}

	return mappedServicesResRequirements, nil
}

func GetFreeResources() (models.FreeResources, error) {
	var freeResource models.FreeResources
	if res := db.DB.Find(&freeResource); res.RowsAffected == 0 {
		return freeResource, res.Error
	}
	return freeResource, nil
}

func GetAvailableResources(gormDB *gorm.DB) (models.AvailableResources, error) {
	var availableRes models.AvailableResources
	if res := gormDB.Find(&availableRes); res.RowsAffected == 0 {
		return availableRes, res.Error
	}
	return availableRes, nil
}
