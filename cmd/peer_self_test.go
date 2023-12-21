package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/models"
)

func setMockMetadata() error {
	// set metadata path
	metadataPath := config.GetConfig().MetadataPath
	metadataFullPath := fmt.Sprintf("%s/metadataV2.json", metadataPath)

	mockMetadataJSON := []byte(`{
        "name": "metadata",
        "resource": {
            "memory_max": 256,
            "total_core": 4,
            "cpu_max": 700
        },
        "available": {
            "cpu": 690,
            "memory": 246
        },
        "reserved": {
            "cpu": 10,
            "memory": 10
        },
        "network": "tcp",
        "public_key": "abc123"
    }`)

	// write mock content inside metadata
	err := afero.WriteFile(mockFS, metadataFullPath, mockMetadataJSON, 0644)
	if err != nil {
		return fmt.Errorf("error writing mock content to file: %v", err)
	}

	return nil
}

func setupMockDB() error {
	mockDB, err := initMockDB()
	if err != nil {
		return fmt.Errorf("failed to initialize mock db: %v", err)
	}

	err = resetMockDB(mockDB, models.Libp2pInfo{})
	if err != nil {
		return fmt.Errorf("failed to reset previous db tables: %v", err)
	}

	err = mockDB.AutoMigrate(&models.Libp2pInfo{})
	if err != nil {
		return fmt.Errorf("unable to auto migrate mock db: %v", err)
	}

	mockP2PInfo := models.Libp2pInfo{
		ID:         1,
		PrivateKey: []byte("secretkey"),
	}

	// insert mocked data inside db
	result := mockDB.Create(&mockP2PInfo)
	if result.Error != nil {
		return fmt.Errorf("failed to insert data inside mock db: %v", err)
	}

	return nil
}

func Test_SelfPeerCmd(t *testing.T) {
	assert := assert.New(t)

	err := setupMockDB()
	assert.NoError(err)

	err = setMockMetadata()
	assert.NoError(err)

	mockUtils := &MockUtilsService{}

	selfResponse := []byte(`{
    "ID": "abcdef12345",
    "Addrs": ["ip4/10000", "ip6/20000"]
    }`)
	mockUtils.SetResponseFor("GET", "/api/v1/peers/self", selfResponse)

	buf := new(bytes.Buffer)
	cmd := NewPeerSelfCmd(mockUtils)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err = cmd.Execute()
	assert.NoError(err)

	buf2 := new(bytes.Buffer)
	fmt.Fprintln(buf2, "Host ID: abcdef12345")
	fmt.Fprintln(buf2, "ip4/10000, ip6/20000")

	assert.Equal(buf.String(), buf2.String())
}
