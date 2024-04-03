package onboarding

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/dms/resources"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/internal/heartbeat"
	"gitlab.com/nunet/device-management-service/internal/klogger"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"

	"github.com/spf13/afero"
)

var FS afero.Fs = afero.NewOsFs()
var AFS *afero.Afero = &afero.Afero{Fs: FS}

// GetMetadata reads metadataV2.json file and returns a models.MetadataV2 struct
func GetMetadata() (*models.Metadata, error) {
	metadataPath := utils.GetMetadataFilePath()
	content, err := AFS.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read metadata file: %w", err)
	}
	var metadata models.Metadata
	err = json.Unmarshal(content, &metadata)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal metadata: %w", err)
	}
	return &metadata, nil
}

func CreatePaymentAddress(wallet string) (*models.BlockchainAddressPrivKey, error) {
	var (
		pair *models.BlockchainAddressPrivKey
		err  error
	)
	if wallet == "ethereum" {
		pair, err = GetEthereumAddressAndPrivateKey()
	} else if wallet == "cardano" {
		pair, err = GetCardanoAddressAndMnemonic()
	} else {
		return nil, fmt.Errorf("invalid wallet")
	}
	if err != nil {
		return nil, fmt.Errorf("could not generate %s address: %w", wallet, err)
	}
	return pair, nil
}

func Status() (*models.OnboardingStatus, error) {
	var (
		metadataPath string
		dbPath       string
	)
	configPath := config.GetConfig().General.MetadataPath
	onboarded, err := utils.IsOnboarded()
	if err != nil {
		return nil, fmt.Errorf("could not check onboard status: %w", err)
	}

	metadataPath = filepath.Join(configPath, "metadataV2.json")
	metadata, err := AFS.Exists(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("unable to check metadata file: %w", err)
	}
	if !metadata {
		metadataPath = ""
	}

	dbPath = filepath.Join(configPath, "nunet.db")
	db, err := AFS.Exists(dbPath)
	if err != nil {
		return nil, fmt.Errorf("unable to check database file: %w", err)
	}
	if !db {
		dbPath = ""
	}

	resp := models.OnboardingStatus{
		Onboarded:    onboarded,
		Error:        err,
		MachineUUID:  utils.GetMachineUUID(),
		MetadataPath: metadataPath,
		DatabasePath: dbPath,
	}
	return &resp, nil
}

func Onboard(ctx context.Context, capacity models.CapacityForNunet) (*models.Metadata, error) {
	configPath := config.GetConfig().General.MetadataPath
	configExist, err := AFS.DirExists(configPath)
	if err != nil {
		return nil, fmt.Errorf("could not check if config directory exists: %w", err)
	}
	if !configExist {
		return nil, fmt.Errorf("config directory does not exist: %w", err)
	}

	hostname, _ := os.Hostname()

	totalCpu := resources.GetTotalProvisioned().CPU
	totalMem := resources.GetTotalProvisioned().Memory
	numCores := resources.GetTotalProvisioned().NumCores

	var metadata models.Metadata
	metadata.Name = hostname
	metadata.UpdateTimestamp = time.Now().Unix()
	metadata.Resource.MemoryMax = int64(totalMem)
	metadata.Resource.TotalCore = int64(numCores)
	metadata.Resource.CPUMax = int64(totalCpu)

	err = utils.ValidateAddress(capacity.PaymentAddress)
	if err != nil {
		return nil, fmt.Errorf("could not validate payment address: %w", err)
	}

	err = validateMemoryAndCPU(capacity.CPU, capacity.Memory)
	if err != nil {
		return nil, fmt.Errorf("could not validate capacity data: %w", err)
	}

	metadata.AllowCardano = false
	if capacity.Cardano {
		if capacity.Memory < 10000 || capacity.CPU < 6000 {
			return nil, fmt.Errorf("cardano node requires 10000MB of RAM and 6000MHz CPU")
		}
		metadata.AllowCardano = true
	}

	gpuInfo, err := resources.Check_gpu()
	if err != nil {
		zlog.Sugar().Errorf("unable to detect GPU: %v ", err.Error())
	}
	metadata.GpuInfo = gpuInfo

	channels := []string{"nunet-staging", "nunet-test", "nunet-team", "nunet-edge"}
	validChannel := utils.SliceContains(channels, capacity.Channel)
	if !validChannel {
		return nil, fmt.Errorf("invalid channel data: '%s' channel does not exist", capacity.Channel)
	}

	metadata.Reserved.Memory = capacity.Memory
	metadata.Reserved.CPU = capacity.CPU
	metadata.Available.Memory = int64(totalMem) - capacity.Memory
	metadata.Available.CPU = int64(totalCpu) - capacity.CPU
	metadata.Network = capacity.Channel
	metadata.PublicKey = capacity.PaymentAddress
	metadata.NTXPricePerMinute = capacity.NTXPricePerMinute

	file, _ := json.MarshalIndent(metadata, "", " ")
	metadataPath := filepath.Join(configPath, "metadataV2.json")
	err = AFS.WriteFile(metadataPath, file, 0644)
	if err != nil {
		return nil, fmt.Errorf("could not write to metadata file: %w", err)
	}

	avalRes := models.AvailableResources{
		TotCpuHz:          int(capacity.CPU),
		CpuNo:             int(numCores),
		CpuHz:             resources.Hz_per_cpu(),
		PriceCpu:          0, // TODO: Get price of CPU
		Ram:               int(capacity.Memory),
		PriceRam:          0, // TODO: Get price of RAM
		Vcpu:              int(math.Floor((float64(capacity.CPU)) / resources.Hz_per_cpu())),
		Disk:              0,
		PriceDisk:         0,
		NTXPricePerMinute: capacity.NTXPricePerMinute,
	}

	var aval models.AvailableResources
	res := db.DB.WithContext(ctx).Find(&aval)
	if res.RowsAffected == 0 {
		res = db.DB.WithContext(ctx).Create(&avalRes)
		if res.Error != nil {
			return nil, fmt.Errorf("unable to create available resources table: %w", res.Error)
		}
	} else {
		res = db.DB.WithContext(ctx).Model(aval).Where("id = ?", 1).Updates(avalRes)
		if res.Error != nil {
			return nil, fmt.Errorf("unable to update available resources table: %w", res.Error)
		}
	}

	priv, pub, err := libp2p.GenerateKey(0)
	if err != nil {
		return nil, fmt.Errorf("unable to generate key pair: %w", err)
	}

	err = libp2p.SaveNodeInfo(priv, pub, capacity.ServerMode, capacity.IsAvailable)
	if err != nil {
		return nil, fmt.Errorf("unable to save node info: %w", err)
	}

	err = resources.CalcFreeResAndUpdateDB()
	if err != nil {
		// JUST LOG ERRORS
		return nil, fmt.Errorf("could not calculate free resources and update database: %w", err)
	}

	hostID, err := libp2p.RegisterRunNewNode(capacity.ServerMode, capacity.IsAvailable)
	if err != nil {
		return nil, fmt.Errorf("could not register and run new node: %w", err)
	}

	_, err = heartbeat.NewToken(hostID, capacity.Channel)
	if err != nil {
		zlog.Sugar().Errorf("unable to get new telemetry token: %v", err)
	}
	klogger.Logger.Info("device onboarded")

	_, err = utils.RegisterLogbin(utils.GetMachineUUID(), hostID)
	if err != nil {
		zlog.Sugar().Errorf("unable to register with logbin: %v", err)
	}
	return &metadata, nil
}

func ResourceConfig(ctx context.Context, resource models.ResourceConfig) (*models.Metadata, error) {
	onboarded, err := utils.IsOnboarded()
	if err != nil {
		return nil, fmt.Errorf("could not check onboard status: %w", err)
	}
	if !onboarded {
		return nil, fmt.Errorf("machine is not onboarded")
	}

	err = validateMemoryAndCPU(resource.CPU, resource.Memory)
	if err != nil {
		return nil, fmt.Errorf("could not validate capacity data: %w", err)
	}

	metadata, err := utils.ReadMetadataFile()
	if err != nil {
		return nil, fmt.Errorf("could not read metadata file: %w", err)
	}
	metadata.Reserved.CPU = resource.CPU
	metadata.Reserved.Memory = resource.Memory
	metadata.NTXPricePerMinute = resource.NTXPricePerMinute

	var aval models.AvailableResources
	res := db.DB.WithContext(ctx).First(&aval)
	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("no rows affected in available resources table")
	}
	aval.TotCpuHz = int(resource.CPU)
	aval.Ram = int(resource.Memory)
	aval.NTXPricePerMinute = resource.NTXPricePerMinute

	db.DB.Save(&aval)

	file, _ := json.MarshalIndent(metadata, "", " ")

	metadataPath := utils.GetMetadataFilePath()
	err = AFS.WriteFile(metadataPath, file, 0644)
	if err != nil {
		return nil, fmt.Errorf("could not write to metadata file: %w", err)
	}
	klogger.Logger.Info("device resource changed")

	err = resources.CalcFreeResAndUpdateDB()
	if err != nil {
		return nil, fmt.Errorf("could not calculate free resources and update database: %w", err)
	}
	return metadata, nil
}

func Offboard(ctx context.Context, force bool) error {
	klogger.Logger.Info("device offboarding started")
	onboarded, err := utils.IsOnboarded()
	if err != nil && !force {
		return fmt.Errorf("could not retrieve onboard status: %w", err)
	} else if err != nil && force {
		zlog.Sugar().Errorf("problem with onboarding state: %v", err)
		zlog.Info("continuing with offboarding because forced")
	}

	if !onboarded {
		return fmt.Errorf("machine is not onboarded")
	}

	err = libp2p.ShutdownNode()
	if err != nil {
		return fmt.Errorf("unable to shutdown node: %w", err)
	}

	metadataPath := utils.GetMetadataFilePath()
	err = os.Remove(metadataPath)
	if err != nil && !force {
		return fmt.Errorf("failed to remove metadata file: %w", err)
	} else if err != nil && force {
		zlog.Sugar().Errorf("failed to delete metadata file - problem with onboarding state: %v", err)
		zlog.Info("continuing with offboarding because forced")
	}

	// delete the available resources from database
	var aval models.AvailableResources
	res := db.DB.WithContext(ctx).Where("id = ?", 1).Delete(&aval)
	if res.Error != nil {
		zlog.Error(res.Error.Error())
	} else if res.RowsAffected == 0 && !force {
		zlog.Error("no rows were affected while deleting available resources")
		return fmt.Errorf("unable to delete available resources on database: %w", err)
	}

	klogger.Logger.Info("device offboarded successfully")
	return nil
}

func fileExists(filename string) bool {
	info, err := AFS.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func validateResource(requested, total int64, resourceType string) error {
	min := total / 10
	max := total * 9 / 10
	if requested < min || requested > max {
		return fmt.Errorf("%s should be between 10%% and 90%% of the available %s (%d and %d)", resourceType, resourceType, min, max)
	}
	return nil
}

func validateMemoryAndCPU(cpu, memory int64) error {
	total := resources.GetTotalProvisioned()

	if err := validateResource(cpu, int64(total.CPU), "CPU"); err != nil {
		return err
	}
	if err := validateResource(memory, int64(total.Memory), "memory"); err != nil {
		return err
	}
	return nil
}