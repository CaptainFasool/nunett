package telemetry

import (
	"reflect"
	"testing"

	"gitlab.com/nunet/device-management-service/models"
)

func getMockHardwareResources() *HardwareResources {
	h := &HardwareResources{
		DBFreeResources: models.FreeResources{
			Resources: models.Resources{
				TotCPU: 5000,
				RAM:    4000,
				Disk:   500,
			},
		},
		OnboardedResources: models.OnboardedResources{
			Resources: models.Resources{
				TotCPU:  10000,
				RAM:     9000,
				Disk:    1000,
				CoreCPU: 1000,
			},
		},
	}
	h.NewFreeRes = h.DBFreeResources
	return h
}

func mockResourcesToModify() models.Resources {
	return models.Resources{
		TotCPU: 1000,
		RAM:    1000,
		Disk:   100,
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
					Resources: models.Resources{
						TotCPU: 6000,
						RAM:    5000,
						Disk:   600,
						VCPU:   6,
					},
				},
				OnboardedResources: initialResources.OnboardedResources,
			},
		},
		{
			name:               "Test Decrease Operation",
			resourcesToMod:     resourcesToChange,
			increaseOrDecrease: -1,
			expected: &HardwareResources{
				DBFreeResources: initialResources.DBFreeResources,
				NewFreeRes: models.FreeResources{
					Resources: models.Resources{
						TotCPU: 4000,
						RAM:    3000,
						Disk:   400,
						VCPU:   4,
					},
				},
				OnboardedResources: initialResources.OnboardedResources,
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

			if !reflect.DeepEqual(hardwareResources.OnboardedResources, tt.expected.OnboardedResources) {
				t.Errorf(
					"Expected OnboardedResources to remain unchanged. Want: %+v, Got %+v",
					tt.expected.OnboardedResources, hardwareResources.OnboardedResources,
				)
			}

		})

	}
}

func TestModifyFreeResourcesMultipleTimes(t *testing.T) {
	initialResources := getMockHardwareResources()
	resourcesToChange := mockResourcesToModify()
	tests := []struct {
		name               string
		resourcesToMod     models.Resources
		increaseOrDecrease int // 1 for increasing, -1 for decreasing
		expected           *HardwareResources
	}{
		{
			name:               "Test Increasing two times",
			resourcesToMod:     resourcesToChange,
			increaseOrDecrease: 1,
			expected: &HardwareResources{
				DBFreeResources: initialResources.DBFreeResources,
				NewFreeRes: models.FreeResources{
					Resources: models.Resources{
						TotCPU: 7000,
						RAM:    6000,
						Disk:   700,
						VCPU:   7,
					},
				},
				OnboardedResources: initialResources.OnboardedResources,
			},
		},
		{
			name:               "Test Decreasing two times",
			resourcesToMod:     resourcesToChange,
			increaseOrDecrease: -1,
			expected: &HardwareResources{
				DBFreeResources: initialResources.DBFreeResources,
				NewFreeRes: models.FreeResources{
					Resources: models.Resources{
						TotCPU: 3000,
						RAM:    2000,
						Disk:   300,
						VCPU:   3,
					},
				},
				OnboardedResources: initialResources.OnboardedResources,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hardwareResources := getMockHardwareResources()
			hardwareResources.modifyFreeResources(tt.resourcesToMod, tt.increaseOrDecrease)
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

			if !reflect.DeepEqual(hardwareResources.OnboardedResources, tt.expected.OnboardedResources) {
				t.Errorf(
					"Expected OnboardedResources to remain unchanged. Want: %+v, Got %+v",
					tt.expected.OnboardedResources, hardwareResources.OnboardedResources,
				)
			}

		})

	}
}
