package plugins

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/aminjam/hipops/utilities"
)

type Plugin interface {
	DefaultPlay() string
	Mask(string) string
	Unmask(string) string
	Run(*Action, *PluginConfig) error
}
type PluginConfig struct {
	GitKey, PrivateKey, PlaybookPath string
	Debug                            int
}
type Action struct {
	Name              string           `json:"name"`
	Dest              string           `json:"dest"`
	Inventory         string           `json:"inventory"`
	User              string           `json:"user"`
	Play              string           `json:"play"`
	PythonInterpreter string           `json:"ansible_python_interpreter,omitempty"`
	Repository        *Repository      `json:"repository"`
	Files             []*Customization `json:"files"`
	Containers        []*Container     `json:"containers"`
}
type Repository struct {
	Branch string `json:"branch"`
	SshUrl string `json:"sshUrl"`
	SshKey string `json:"sshKey"`
	Folder string `json:"folder"`
}

func (r *Repository) Configure() (err error) {
	if strings.Contains(r.SshUrl, "@") || !strings.Contains(r.SshUrl, ".git") {
		err = errors.New(utilities.INVALID_REPOSITORY)
	}
	if r.Branch == "" {
		r.Branch = utilities.DEFAULT_APP_BRANCH
	}
	return
}

type Customization struct {
	Src        string `json:"src"`
	Dest       string `json:"dest"`
	DestFolder string `json:"destFolder"`
	Mode       int    `json:"mode"`
}

func (c *Customization) Configure(suffix string, appDest string) (err error) {
	if strings.HasPrefix(c.Src, "http") {
		c.Src, err = utilities.DownloadFile(c.Src, suffix)
		if err != nil {
			return
		}
	} else if !strings.HasPrefix(c.Src, "/") {
		c.Src = "@BASEDIR/" + c.Src
	}
	if c.Mode == 0 {
		c.Mode = 400
	}
	if !strings.HasPrefix(c.Dest, "~") || !strings.HasPrefix(c.Dest, "/") {
		c.Dest = fmt.Sprintf("%s/%s", appDest, c.Dest)
	}
	c.DestFolder = c.Dest[:strings.LastIndex(c.Dest, "/")]
	return nil
}

type Container struct {
	Params string `json:"params"`
	Name   string `json:"name"`
	State  string `json:"state"`
}

func (c *Container) Configure() {
	var p = regexp.MustCompile(`--name\s.+\s`)
	match := p.FindString(c.Params)
	if match != "" {
		c.Name = strings.Split(p.FindString(c.Params), " ")[1]
	} else {
		c.Params = fmt.Sprintf("--name %s %s", c.Name, c.Params)
	}
}
