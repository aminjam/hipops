package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/aminjam/hipops/utilities"
)

type os struct {
	User, PythonInterpreter string
}

type app struct {
	Host, Image, Name,
	Dest, Type string
	Ports          []int
	Cred           *cred
	Customizations []*customization
	Repository     *repository
}

func (a *app) configure(sc *Scenario) error {
	if a.Type == "" {
		a.Type = utilities.DEFAULT_APP_TYPE
	}
	if a.Type != utilities.DEFAULT_APP_TYPE {
		a.Name = fmt.Sprintf("%s-%s-%s", sc.Id, a.Type, a.Name)
	}
	if a.Dest == "" {
		a.Dest = fmt.Sprintf("%s/%s-%s/%s/%s", sc.Dest, sc.Id, sc.Env, a.Type, a.Name)
	}
	a.Dest = strings.TrimSuffix(a.Dest, "/")
	for c, _ := range a.Customizations {
		if err := a.Customizations[c].configure(sc.Suffix, a.Dest); err != nil {
			return err
		}
	}
	if a.Repository != nil {
		if err := a.Repository.configure(); err != nil {
			return err
		}
	}
	return nil
}

type repository struct {
	Branch string `json:"branch"`
	SshUrl string `json:"sshUrl"`
	SshKey string `json:"sshKey"`
	Folder string `json:"folder"`
}

func (r *repository) configure() (err error) {
	if strings.Contains(r.SshUrl, "@") || !strings.Contains(r.SshUrl, ".git") {
		err = errors.New(utilities.INVALID_REPOSITORY)
	}
	if r.Branch == "" {
		r.Branch = utilities.DEFAULT_APP_BRANCH
	}
	return
}

type customization struct {
	Src        string `json:"src"`
	Dest       string `json:"dest"`
	DestFolder string `json:"destFolder"`
	Mode       int    `json:"mode"`
}

func (c *customization) configure(suffix string, appDest string) (err error) {
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

type cred struct {
	DbName, Username, Password string
}

type playbook struct {
	Name, Play, State,
	Inventory, User string
	Actions []*docker
	Apps    []string
}

type docker struct {
	Image  string `json:"image"`
	Params string `json:"params"`
	Name   string `json:"name"`
	State  string `json:"state"`
}

type Scenario struct {
	Env, Id, Description,
	Dest, Suffix string
	Oses      []*os
	Apps      []*app
	Playbooks []*playbook
}

func (sc *Scenario) Parse(config []byte) ([]*Action, error) {
	err := json.Unmarshal(config, sc)
	if err != nil {
		return nil, err
	}
	sc.Suffix = fmt.Sprintf("%s-%s", sc.Id, sc.Env)
	if sc.Dest == "" {
		return nil, errors.New(utilities.UNKNOWN_SCENARIO_DEST)
	}
	sc.Dest = strings.TrimSuffix(sc.Dest, "/")

	for i, _ := range sc.Apps {
		if err = sc.Apps[i].configure(sc); err != nil {
			return nil, err
		}
	}

	actions, counter := make([]*Action, sc.countActions()), 0

	for _, p := range sc.Playbooks {
		action := &Action{}
		if err = action.configureOS(sc.Oses, p.User); err != nil {
			return nil, err
		}
		if p.State == "" {
			p.State = utilities.DEFAULT_APP_STATE
		}

		if len(p.Apps) != 0 {
			for _, appString := range p.Apps {
				p.Name = utilities.ParseTemplate("{{.App.Name}}", sc, appString)
				app, err := sc.findApp(p.Name)
				if err != nil {
					return nil, err
				}
				if err = action.configureApp(app); err != nil {
					return nil, err
				}
			}
		}
		action.Play = p.Play
		fmt.Println("USER", action.User, action.PythonInterpreter)
		actions[counter] = action
		counter++
	}

	return actions, nil
}

func (sc *Scenario) findApp(name string) (*app, error) {
	if name != "" {
		for k, v := range sc.Apps {
			if strings.Contains(v.Name, name) {
				sc.Apps[k].configure(sc)
				return sc.Apps[k], nil
			}
		}
	}
	return nil, errors.New(utilities.APP_NOT_FOUND)
}
func (sc *Scenario) countActions() int {
	total := 0
	for _, p := range sc.Playbooks {
		count := len(p.Apps)
		if count <= 0 {
			count = 1
		}
		total += count
	}
	return total
}
