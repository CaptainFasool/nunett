package firecracker

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/models"
)

func RunPreviouslyRunningVMs() error {
	var vms []models.VirtualMachine

	if result := db.DB.Where("state = ?", "running").Find(&vms); result.Error != nil {
		panic(result.Error)
	}

	for _, vm := range vms {
		// send request to fromConfig
		jsonBytes, _ := json.Marshal(vm)
		// MakeInternalRequest(&gin.Context{}, "POST", "/vm/fromConfig", jsonBytes)

		// set the HTTP method, url, and request body
		req, _ := http.NewRequest("POST", DMS_BASE_URL+"/vm/fromConfig", bytes.NewBuffer(jsonBytes))

		client := http.Client{}
		// set the request header Content-Type for json
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		req.Header.Set("Accept", "application/json")
		_, err := client.Do(req)
		if err != nil {
			// c.JSON(400, gin.H{
			// 	"message":   errMsg,
			// 	"timestamp": time.Now(),
			// })
			// return
			panic(err)
		}
	}
	return nil
}

func GenerateSocketFile(n int) string {
	prefix := "/etc/nunet/sockets/"
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	rand.Seed(time.Now().Unix())

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return prefix + string(s) + ".socket"
}
