package utilities

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/mitchellh/cli"
)

const (
	DEFAULT_APP_STATE  = "running"
	DEFAULT_APP_TYPE   = "generic"
	DEFAULT_APP_BRANCH = "master"
)

func CheckErr(err error) {
	ui := &cli.BasicUi{Writer: os.Stdout}
	if err != nil {
		msg := fmt.Sprintf("%s", err)
		ui.Error(msg)
		log.Fatalf(msg)
	}
}
func ParseTemplate(input string, base interface{}, app string) string {
	t := template.New("")
	if app != "" {
		input = formatTemplate(input, app)
	}
	t, _ = t.Parse(input)
	buf := new(bytes.Buffer)
	t.Execute(buf, base)
	return buf.String()
}

func formatTemplate(input string, app string) string {
	app = strings.Replace(app, "{{", "(", -1)
	app = strings.Replace(app, "}}", ")", -1)
	var re = regexp.MustCompile(`({{.App(.&}})*)`)
	input = re.ReplaceAllString(input, fmt.Sprintf("{{%s${2}", app))
	re = regexp.MustCompile(`({{index .App.Ports(.&}})*)`)
	input = re.ReplaceAllString(input, fmt.Sprintf("{{index (%s.Ports)${2}", app))
	return input
}
