package utilities

import (
	"fmt"
	"github.com/mitchellh/cli"
	"log"
	"os"
)

func CheckErr(err error) {
	ui := &cli.BasicUi{Writer: os.Stdout}
	if err != nil {
		msg := fmt.Sprintf("%s", err)
		ui.Error(msg)
		log.Fatalf(msg)
	}
}
