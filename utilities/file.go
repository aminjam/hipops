package utilities

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

func CleanupTempFiles(suffix string) {
	d, err := os.Open("/tmp")
	defer d.Close()
	CheckErr(err)

	files, err := d.Readdir(-1)
	CheckErr(err)

	fmt.Println("Reading files for /tmp")

	for _, file := range files {
		if file.Mode().IsRegular() {
			if strings.HasPrefix(file.Name(), "hipops-"+suffix) {
				os.Remove("/tmp/" + file.Name())
				fmt.Println("Deleted ", file.Name())
			}
		}
	}
}

func DownloadFile(url string, suffix string) (string, error) {
	rand.Seed(time.Now().UnixNano())
	fileName := fmt.Sprintf("/tmp/hipops-%s-%v", suffix, rand.Intn(1000000))
	fmt.Println("Downloading file...")

	output, err := os.Create(fileName)
	defer output.Close()

	response, err := http.Get(url)
	defer response.Body.Close()
	if err != nil {
		return "", err
	}

	_, err = io.Copy(output, response.Body)
	if err != nil {
		return "", err
	}
	return fileName, nil
}
