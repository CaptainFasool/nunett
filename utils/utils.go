package utils

import (
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/models"
)

var KernelFileURL = "https://d.nunet.io/fc/vmlinux"
var KernelFilePath = "/etc/nunet/vmlinux"
var FilesystemURL = "https://d.nunet.io/fc/nunet-fc-ubuntu-20.04-0.ext4"
var FilesystemPath = "/etc/nunet/nunet-fc-ubuntu-20.04-0.ext4"

func DownloadFile(url string, filepath string) (err error) {
	zlog.Sugar().Infof("Downloading file '", filepath, "' from '", url, "'")
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	log.Println("Finished downloading file '", filepath, "'")
	return nil
}

func RandomString(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	sb := strings.Builder{}
	sb.Grow(n)
	for i := 0; i < n; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
}

func GetChannelName() string {
	metadata, err := ReadMetadataFile()
	if err != nil {
		zlog.Sugar().Errorf("could not read metadata: %v", err)
	}
	return metadata.Network
}

func GenerateMachineUUID() (string, error) {
	var machine models.MachineUUID
	machineUUID, err := uuid.NewDCEGroup()
	if err != nil {
		return "", err
	}
	machine.UUID = machineUUID.String()

	result := db.DB.Create(&machine)
	if result.Error != nil {
		return "", result.Error
	}

	return machine.UUID, nil
}

func GetMachineUUID() string {
	var machine models.MachineUUID

	// try db
	result := db.DB.First(&machine)
	if result.Error == nil {
		if machine.UUID != "" {
			return machine.UUID
		}
	}

	return machine.UUID

}

// ReadMetadata returns metadata from metadataV2.json file
func ReadMetadataFile() (models.MetadataV2, error) {
	metadataF, err := os.ReadFile("/etc/nunet/metadataV2.json")
	if err != nil {
		return models.MetadataV2{}, err
	}
	var metadata models.MetadataV2
	err = json.Unmarshal(metadataF, &metadata)
	if err != nil {
		return models.MetadataV2{}, err
	}
	return metadata, nil
}
