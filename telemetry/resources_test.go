package telemetry

import (
	"reflect"
	"testing"

	"gitlab.com/nunet/device-management-service/models"
)

func getMockHardwareResources() *HardwareResources {
	h := &HardwareResources{
		DBFreeResources: models.FreeResources{
			TotCpuHz: 5000,
			Ram:      4000,
			Disk:     500,
		},
		AvailableResources: models.AvailableResources{
			TotCpuHz: 10000,
			Ram:      9000,
			Disk:     1000,
			CpuHz:    1000,
		},
	}
	h.NewFreeRes = h.DBFreeResources
	return h
}

func mockResourcesToModify() models.Resources {
	return models.Resources{
		TotCpuHz: 1000,
		Ram:      1000,
		Disk:     100,
	}
}

func TestModifyFreeResources(t *testing.T) {
	initialResources := getMockHardwareResources()
	resourcesToChange := mockResourcesToModify()
	tests := []struct {
		name               string
		resourcesToMod     models.Resources
		increaseOrDecrease int // 1 for increasing, -1 for decreasing
		expected           *HardwareResources
	}{
		{
			name:               "Test Increase Operation",
			resourcesToMod:     resourcesToChange,
			increaseOrDecrease: 1,
			expected: &HardwareResources{
				DBFreeResources: initialResources.DBFreeResources,
				NewFreeRes: models.FreeResources{
					TotCpuHz: 6000,
					Ram:      5000,
					Disk:     600,
					Vcpu:     6,
				},
				AvailableResources: initialResources.AvailableResources,
			},
		},
		{
			name:               "Test Decrease Operation",
			resourcesToMod:     resourcesToChange,
			increaseOrDecrease: -1,
			expected: &HardwareResources{
				DBFreeResources: initialResources.DBFreeResources,
				NewFreeRes: models.FreeResources{
					TotCpuHz: 4000,
					Ram:      3000,
					Disk:     400,
					Vcpu:     4,
				},
				AvailableResources: initialResources.AvailableResources,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hardwareResources := getMockHardwareResources()
			hardwareResources.modifyFreeResources(tt.resourcesToMod, tt.increaseOrDecrease)

			if !reflect.DeepEqual(hardwareResources.NewFreeRes, tt.expected.NewFreeRes) {
				t.Errorf(
					"Expected NewFreeRes to be %+v, but got %+v",
					tt.expected.NewFreeRes, hardwareResources.NewFreeRes,
				)
			}

			if !reflect.DeepEqual(hardwareResources.DBFreeResources, tt.expected.DBFreeResources) {
				t.Errorf(
					"Expected DBFreeResources to remain unchanged. Want: %+v, Got %+v",
					tt.expected.DBFreeResources, hardwareResources.DBFreeResources,
				)
			}

			if !reflect.DeepEqual(hardwareResources.AvailableResources, tt.expected.AvailableResources) {
				t.Errorf(
					"Expected AvailableResources to remain unchanged. Want: %+v, Got %+v",
					tt.expected.AvailableResources, hardwareResources.AvailableResources,
				)
			}

		})
	}
}
