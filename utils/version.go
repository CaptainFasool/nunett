package utils

import (
	"fmt"
	"io"

	"github.com/buger/jsonparser"
)

func GetDMSVersion() (string) {
	// above insert a function that checks if dms is running
	resp, err := MakeInternalRequest(nil, "GET", "/swagger/doc.json", nil)
	if err != nil {
		return "DMS has not been initialized. Run 'nunet daemon'"
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	ver, err := jsonparser.GetString(body, "info", "version")
	if err != nil {
		fmt.Println(err)
	}

	return ver
}


