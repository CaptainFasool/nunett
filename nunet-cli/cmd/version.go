/*
Copyright © 2023 Gustavo Silva <gustavo.silva@nunet.io>
*/

package cmd

import (
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/spf13/cobra"
	"io"
	"log"
	"net/http"
	"os"
)

var swaggerURL = "http://localhost:7777/swagger/doc.json"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show DMS version",
	Long:  "Prints to the user the current DMS version",
	Run: func(cmd *cobra.Command, args []string) {
		req, err := http.NewRequest("GET", swaggerURL, nil)
		if err != nil {
			log.Fatal(err)
		} // Create request

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		} // Client make request

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		} // Read body response

		data, err := jsonparser.GetString(body, "info", "version")
		if err != nil {
			log.Fatal(err)
		} // Parse JSON and select field

		v := fmt.Sprintf("nunet-dms v%s\n", data)
		if _, err := io.WriteString(os.Stdout, v); err != nil {
			log.Fatal(err)
		} // Print to stdout
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
