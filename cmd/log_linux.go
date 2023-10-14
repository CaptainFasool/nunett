//go:build linux

package cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/coreos/go-systemd/sdjournal"
	"github.com/spf13/cobra"
)

func init() {

}

var logCmd = &cobra.Command{
	Use:    "log",
	Short:  "Gather all logs into a tarball",
	Long:   "",
	PreRun: isDMSRunning(),
	Run: func(cmd *cobra.Command, args []string) {
		logDir := "/tmp/nunet-log"
		dmsLogDir := filepath.Join(logDir, "dms-log")

		fmt.Println("Collecting logs...")

		// create directory with necessary permissions
		err := os.MkdirAll(dmsLogDir, 0777)
		if err != nil {
			fmt.Println("Error creating directory:", err)
			os.Exit(1)
		}

		// initialize journal
		j, err := sdjournal.NewJournal()
		if err != nil {
			fmt.Println("Error initializing journal:", err)
			os.Exit(1)
		}
		defer j.Close()

		// filter by service unit name
		if err := j.AddMatch("_SYSTEMD_UNIT=nunet-dms.service"); err != nil {
			fmt.Println("Error adding match:", err)
			os.Exit(1)
		}

		bootIDs := map[string]int{}

		// loop through journal entries
		for {
			c, err := j.Next()
			if err != nil {
				fmt.Println("Error reading next journal entry:", err)
				continue
			}

			if c == 0 {
				break
			}

			entry, err := j.GetEntry()
			if err != nil {
				fmt.Println("Error getting journal entry:", err)
				continue
			}

			bootID := entry.Fields["_BOOT_ID"]
			if _, exists := bootIDs[bootID]; !exists {
				bootIDs[bootID] = len(bootIDs) + 1
			}

			logData := fmt.Sprintf("%s: %s\n", entry.RealtimeTimestamp, entry.Fields["MESSAGE"])

			logFilePath := filepath.Join(dmsLogDir, fmt.Sprintf("dms_log.%d", bootIDs[bootID]))
			if err := appendToFile(logFilePath, logData); err != nil {
				fmt.Printf("Error writing log file for boot %d: %s", bootIDs[bootID], err)
			}
		}

		// create tar archive
		tarGzFile := filepath.Join(logDir, "nunet-log.tar.gz")
		if err := createTar(tarGzFile, dmsLogDir); err != nil {
			fmt.Println("Error creating tar archive:", err)
			os.Exit(1)
		}

		// remove dms-log directory
		if err := os.RemoveAll(dmsLogDir); err != nil {
			fmt.Println("Error removing dms-log directory: %s", err)
			os.Exit(1)
		}

		fmt.Println(tarGzFile)
	},
}

func appendToFile(filename, data string) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(data); err != nil {
		return err
	}

	return nil
}

func createTar(tarGzPath string, sourceDir string) error {
	tarGzFile, err := os.Create(tarGzPath)
	if err != nil {
		return err
	}
	defer tarGzFile.Close()

	gzWriter := gzip.NewWriter(tarGzFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == tarGzPath {
			return nil
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(path, sourceDir)
		if header.Name == "" || header.Name == "/" {
			return nil
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if info.Mode().IsRegular() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if _, err := tarWriter.Write(data); err != nil {
				return err
			}
		}
		return nil
	})
}
