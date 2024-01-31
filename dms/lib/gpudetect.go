package library

import (
	"strings"

	"github.com/jaypipes/ghw"
)

type GPUVendor int

const (
	Unknown GPUVendor = iota
	NVIDIA
	AMD
)

func (g GPUVendor) String() string {
	switch g {
	case Unknown:
		return "Unknown"
	case NVIDIA:
		return "NVIDIA"
	case AMD:
		return "AMD"
	default:
		return "Unknown"
	}
}

func DetectGPUVendors() ([]GPUVendor, error) {
	var vendors []GPUVendor
	gpu, err := ghw.GPU()
	if err != nil {
		return nil, err
	}

	for _, card := range gpu.GraphicsCards {
		deviceInfo := card.DeviceInfo
		if deviceInfo != nil {
			class := deviceInfo.Class
			if class != nil {
				className := strings.ToLower(class.Name)
				if strings.Contains(className, "display controller") ||
					strings.Contains(className, "vga compatible controller") ||
					strings.Contains(className, "3d controller") ||
					strings.Contains(className, "2d controller") {
					vendor := card.DeviceInfo.Vendor
					if vendor != nil {
						if strings.Contains(strings.ToLower(vendor.Name), "nvidia") {
							vendors = append(vendors, NVIDIA)
						}
						if strings.Contains(strings.ToLower(vendor.Name), "amd") {
							vendors = append(vendors, AMD)
						}
					}
				}
			}
		}
	}

	if len(vendors) == 0 {
		return []GPUVendor{Unknown}, nil
	}

	return vendors, nil
}
