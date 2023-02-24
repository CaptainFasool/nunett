package utils

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
)

var KernelFileURL = "https://d.nunet.io/fc/vmlinux"
var KernelFilePath = "/etc/nunet/vmlinux"
var FilesystemURL = "https://d.nunet.io/fc/nunet-fc-ubuntu-20.04-0.ext4"
var FilesystemPath = "/etc/nunet/nunet-fc-ubuntu-20.04-0.ext4"

func DownloadFile(url string, filepath string) (err error) {
	log.Println("Downloading file '", filepath, "' from '", url, "'")
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
