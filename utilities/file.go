package utilities

import (
	"fmt"
	"os"
	"strings"
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
