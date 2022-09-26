package trigger_cardano
import (
    "log"
    "net/http"
    "io"
    "fmt"
    "os"
    "crypto/sha256"
    "errors"
    "io/ioutil"
    "strings"
    "github.com/gin-gonic/gin"
    "gitlab.com/nunet/firecracker-images/image-creator/rootfs"
    "gitlab.com/nunet/device-management-service/firecracker"
)


func ProcessDeployement(kernel string, fs string){
    url := firecracker.DMS_BASE_URL+"vm/start-default"
    method := "POST"
    payload := fmt.Sprintf(`{
        "kernel_image_path":"%s",
        "filesystem_path": "%s"}`,kernel, fs)
            log.Printf(payload)
    body := strings.NewReader(payload)
    client := &http.Client {
    }
    req, err := http.NewRequest(method, url, body)

    if err != nil {
        fmt.Println(err)
        return
    }
    req.Header.Add("Content-Type", "application/json")

    res, err := client.Do(req)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer res.Body.Close()

    response, err := ioutil.ReadAll(res.Body)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println(string(response))
}

func Exists(name string) (bool, error) {
    _, err := os.Stat(name)
    if err == nil {
        return true, nil
    }
    if errors.Is(err, os.ErrNotExist) {
        return false, nil
    }
    return false, err
}

func Manual() {}

func CreateFS(size int) string{
	mountDirectory := "/tmp/rootfs"
	outputFolder := "/etc/nunet/cardano"
	imageFile := outputFolder + "/" + "image.ext4"

	rootfs.PrepareOutputFolder(outputFolder)
	rootfs.CreateEmptyImage(size, imageFile)
	rootfs.MountImage(mountDirectory, imageFile)
	rootfs.UnMountImage(mountDirectory)
	rootfs.CheckImageFilesystem(imageFile)
    return imageFile
}

func Downloader(url string ) string {
	segments := strings.Split(url, "/")
	fileName := segments[len(segments)-1]
    fileName = firecracker.FIRECRACKER_KERNEL_LOCATION+fileName
    log.Printf(fileName)
    
    exists ,_ := Exists(fileName)
    if exists == true {
        return fileName
        } 
        
    out, _ := os.Create(fileName)
    defer out.Close()

    log.Printf("downloading kernel image")

    resp, _ := http.Get(url)
    defer resp.Body.Close()

    io.Copy(out, resp.Body)
    log.Printf("downloaded kernel image")
    return fileName
}


func VerifyChecksum(path string) (response string, err error)  {
    // get checksum from the server
    server, err := http.Get(firecracker.FIRECRACKER_KERNEL_CHECKSUM)
    if err != nil {
        return "unable to get kernal checksum", err
    }
    defer server.Body.Close()
    bodyBytes, err := io.ReadAll(server.Body)
        if err != nil {
            return "unable to get kernal checksum", err
        }
    bodyString := string(bodyBytes)

    // get checksum from local img
    f, err := os.Open(path)
    if err != nil {
        return "unable to get kernal checksum", err
    }
    defer f.Close()

    h := sha256.New()
    if _, err := io.Copy(h, f); err != nil {
        return "unable to get kernal checksum", err
    }

    file_checksum := fmt.Sprintf("%x",h.Sum(nil))

    //verify the checksum
    log.Printf("------verifying checksum of the kernel----")
    if (file_checksum == bodyString){
        return "Checksum ok", nil
    }else {
        return  "Checksum verification failed", errors.New("failed")
    }
}

func Auto() (response string, err error) {
    file := Downloader(firecracker.FIRECRACKER_KERNEL)
    verify_checksum, err := VerifyChecksum(file) 
    if err != nil{
        return verify_checksum, err
    }
    log.Printf("---------checksum verified---------")
    return file, nil
}

func PrepareConfig(c *gin.Context)  {
    typed := getDeployeentType(c)
    var kernel string
    var fs string
    if typed == "auto" {
        kernel_file, err := Auto()
        if err != nil{
            log.Printf(kernel_file)
        }
        kernel = kernel_file
        fs = CreateFS(5)
    }

    if typed == "manual" {
        Manual()
        fs = CreateFS(5)
    }
    ProcessDeployement(kernel, fs)
}

func getDeployeentType(c *gin.Context) string {
    path := c.Request.URL.Path
    path_array := strings.Split(path, "/")
    return path_array[len(path_array)-1]
}

func Deploy(c *gin.Context) {
    PrepareConfig(c)
    var data map[string]interface{}
    err := c.BindJSON(&data)
    if err != nil{
        c.AbortWithError(http.StatusBadRequest, err)
        return
        } else{
        c.JSON(200, &data)
    }
}
