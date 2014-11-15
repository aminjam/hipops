package utilities

import (
	"errors"
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

	fmt.Println(fmt.Sprintf("Reading files for /tmp/hipops-%s*", suffix))

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
func WriteFile(content []byte, fileType string, suffix string) (string, error) {
	rand.Seed(time.Now().UnixNano())
	fileName := fmt.Sprintf("/tmp/hipops-%s-%v.%s", suffix, rand.Intn(1000000), fileType)
	output, err := os.Create(fileName)
	defer output.Close()
	if err != nil {
		return "", err
	}
	_, err = io.WriteString(output, fmt.Sprintf("%s", content))
	if err != nil {
		return "", err
	}
	return fileName, nil
}
func Exists(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		return errors.New("")
	}
	return err
}
