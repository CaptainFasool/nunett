package utils

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	gonet "github.com/shirou/gopsutil/net"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/models"
)

var KernelFileURL = "https://d.nunet.io/fc/vmlinux"
var KernelFilePath = "/etc/nunet/vmlinux"
var FilesystemURL = "https://d.nunet.io/fc/nunet-fc-ubuntu-20.04-0.ext4"
var FilesystemPath = "/etc/nunet/nunet-fc-ubuntu-20.04-0.ext4"

// DownloadFile downloads a file from a url and saves it to a filepath
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

// ReadHttpString GET request to http endpoint and return response as string
func ReadHttpString(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(respBody), nil
}

// RandomString generates a random string of length n
func RandomString(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	sb := strings.Builder{}
	sb.Grow(n)
	for i := 0; i < n; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
}

// GetChannelName returns the channel name from the metadata file
func GetChannelName() string {
	metadata, err := ReadMetadataFile()
	if err != nil || metadata.Network == "" {
		return "" // nunet not onboarded or something wrong with metadata file
	}
	return metadata.Network
}

func GetDashboard() string {
	metadata, err := ReadMetadataFile()
	if err != nil || metadata.Dashboard == "" {
		return ""
	}
	return metadata.Dashboard
}

// GenerateMachineUUID generates a machine uuid
func GenerateMachineUUID() (string, error) {
	var machine models.MachineUUID

	machineUUID, err := uuid.NewDCEGroup()
	if err != nil {
		return "", err
	}
	machine.UUID = machineUUID.String()

	return machine.UUID, nil
}

// GetMachineUUID returns the machine uuid from the DB
func GetMachineUUID() string {
	var machine models.MachineUUID
	uuid, err := GenerateMachineUUID()
	if err != nil {
		zlog.Sugar().Errorf("could not generate machine uuid: %v", err)
	}

	machine.UUID = uuid

	result := db.DB.FirstOrCreate(&machine)
	if result.Error != nil {
		zlog.Sugar().Errorf("could not find or create machine uuid record in DB: %v", result.Error)
	}
	return machine.UUID

}

// SliceContains checks if a string exists in a slice
func SliceContains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// RegisterLogbin registers the device with logbin
func RegisterLogbin(uuid string, peer_id string) (string, error) {
	logbinAuthReq := struct {
		PeerID      string `json:"peer_id"`
		MachineUUID string `json:"machine_uuid"`
	}{
		PeerID:      peer_id,
		MachineUUID: uuid,
	}

	jsonAuth, err := json.Marshal(logbinAuthReq)
	if err != nil {
		zlog.Sugar().Errorf("unable to marshal logbin auth request: %v", err)
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, "https://log.nunet.io/api/v1/auth/register", bytes.NewBuffer(jsonAuth))

	if err != nil {
		zlog.Sugar().Errorf("unable to create http request: %v", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		zlog.Sugar().Errorf("unable to register with logbin: %v", err)
		return "", err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		zlog.Sugar().Errorf("unable to read response body from logbin: %v", err)
		return "", err
	}

	tokenResp := struct {
		Token string `json:"token"`
	}{}

	err = json.Unmarshal(respBody, &tokenResp)
	if err != nil {
		zlog.Sugar().Errorf("unable to unmarshal token response: %v", err)
		return "", err
	}

	logbinAuth := models.LogBinAuth{
		Token:       tokenResp.Token,
		PeerID:      peer_id,
		MachineUUID: uuid,
	}
	result := db.DB.FirstOrCreate(&logbinAuth)
	if result.Error != nil {
		zlog.Sugar().Errorf("unable to create logbin auth record in DB: %v", result.Error)
		return "", result.Error
	}
	return logbinAuth.Token, nil
}

// GetLogbinToken returns the logbin token from the DB
func GetLogbinToken() (string, error) {
	var logbinAuth models.LogBinAuth
	result := db.DB.Find(&logbinAuth)
	if result.Error != nil {
		zlog.Sugar().Errorf("unable to find logbin auth record in DB: %v", result.Error)
		return "", result.Error
	}
	return logbinAuth.Token, nil
}

// ReadMetadata returns metadata from metadataV2.json file
func ReadMetadataFile() (*models.MetadataV2, error) {
	metadataF, err := os.ReadFile(fmt.Sprintf("%s/metadataV2.json", config.GetConfig().General.MetadataPath))
	if err != nil {
		return &models.MetadataV2{}, err
	}
	var metadata models.MetadataV2
	err = json.Unmarshal(metadataF, &metadata)
	if err != nil {
		return &models.MetadataV2{}, err
	}
	return &metadata, nil
}

// IsOnboarded checks if the device is onboarded
func IsOnboarded() (bool, error) {
	var libp2pInfo models.Libp2pInfo
	_ = db.DB.Where("id = ?", 1).Find(&libp2pInfo)
	_, err := ReadMetadataFile()

	if err == nil && libp2pInfo.PrivateKey != nil {
		return true, nil
	} else if err != nil && libp2pInfo.PrivateKey == nil {
		return false, nil
	} else {
		return false, err
	}
}

// ReadyForElastic checks if the device is ready to send logs to elastic
func ReadyForElastic() bool {
	elasticToken := models.ElasticToken{}
	db.DB.Find(&elasticToken)
	return elasticToken.NodeId != "" && elasticToken.ChannelName != ""
}

// CreateDirectoryIfNotExists creates a directory if it does not exist
func CreateDirectoryIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

// CalculateSHA256Checksum calculates the SHA256 checksum of a file
func CalculateSHA256Checksum(filePath string) (string, error) {
	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Create a new SHA-256 hash
	hash := sha256.New()

	// Copy the file's contents into the hash object
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	// Calculate the checksum and return it as a hexadecimal string
	checksum := hex.EncodeToString(hash.Sum(nil))
	return checksum, nil
}

// PromptYesNo prompts the user on stdout for a yes or no response on stdin
func PromptYesNo(prompt string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)

	verifyInput := func(input string) bool {
		lowerInput := strings.ToLower(input)
		return lowerInput == "y" || lowerInput == "yes" || lowerInput == "n" || lowerInput == "no"
	}

	for {
		fmt.Print(prompt + ": ")
		response, err := reader.ReadString('\n')

		if err != nil {
			return false, fmt.Errorf("Error reading from buffer: %v", err)
		}

		response = strings.TrimSpace(response)

		if verifyInput(response) {
			lowerResponse := strings.ToLower(response)
			return lowerResponse == "y" || lowerResponse == "yes", nil
		} else {
			fmt.Println("Invalid input. Please enter 'y' or 'n'")
		}
	}
}

// CheckWSL check if running in WSL
func CheckWSL() (bool, error) {
	file, err := os.Open("/proc/version")
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Microsoft") || strings.Contains(line, "WSL") {
			return true, nil
		}
	}

	if scanner.Err() != nil {
		return false, scanner.Err()
	}

	return false, nil
}

// ListenDMSPort check if DMS port is being used
func ListenDMSPort() (bool, error) {
	port := config.GetConfig().Rest.Port

	conns, err := gonet.Connections("all")
	if err != nil {
		return false, err
	}

	for _, conn := range conns {
		if conn.Status == "LISTEN" && uint32(port) == conn.Laddr.Port {
			return true, nil
		}
	}

	return false, nil
}

// SaveServiceInfo updates service info into SP's DMS for claim Reward by SP user
func SaveServiceInfo(cpService models.Services) error {

	var spService models.Services
	err := db.DB.Model(&models.Services{}).Where("tx_hash = ?", cpService.TxHash).Find(&spService).Error
	if err != nil {
		return fmt.Errorf("Unable to find service on SP side: %v", err)
	}
	cpService.ID = spService.ID
	cpService.CreatedAt = spService.CreatedAt

	result := db.DB.Model(&models.Services{}).Where("tx_hash = ?", cpService.TxHash).Updates(&cpService)
	if result.Error != nil {
		return fmt.Errorf("Unable to update service info on SP side: %v", result.Error.Error())
	}

	return nil
}
