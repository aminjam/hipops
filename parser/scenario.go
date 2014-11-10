package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aminjam/hipops/utilities"
	"strings"
)

type os struct {
	User, PythonInterpreter string
}

type app struct {
	Host, Image, Name,
	Dest, Type string
	Ports          []int
	Cred           cred
	Customizations []customization
	Repository     repository
}

func (a *app) setName(sc *Scenario) {
	if a.Type != "global" {
		a.Name = fmt.Sprintf("%s-%s-%s-%s", sc.Id, sc.Env, a.Type, a.Name)
	}
}

type repository struct {
	Branch string `json:"branch"`
	SshUrl string `json:"sshUrl"`
	SshKey string `json:"sshKey"`
	Folder string `json:"folder"`
}

type customization struct {
	Src        string `json:"src"`
	Dest       string `json:"dest"`
	DestFolder string `json:"destFolder"`
	Mode       int    `json:"mode"`
}

type cred struct {
	DbName, Username, Password string
}

type playbook struct {
	Name, Play, State,
	Inventory, User string
	Actions []docker
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
	Apps      []app
	Playbooks []playbook
}

func (sc *Scenario) Parse(config []byte) ([]*Action, error) {
	err := json.Unmarshal(config, sc)
	if err != nil {
		return nil, err
	}
	sc.Suffix = fmt.Sprintf("%s-%s", sc.Id, sc.Env)
	sc.Dest = strings.TrimSuffix(sc.Dest, "/")

	for i, _ := range sc.Apps {
		sc.Apps[i].setName(sc)
	}

	actions, counter := make([]*Action, sc.countActions()), 0

	for _, p := range sc.Playbooks {
		action := &Action{}
		if err = action.setUser(sc.Oses, p.User); err != nil {
			return nil, err
		}
		fmt.Println("USER", action.User, action.PythonInterpreter)
		actions[counter] = action
		counter++
	}

	return actions, nil
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

type Action struct {
	Dest              string          `json:"dest"`
	Inventory         string          `json:"inventory"`
	User              string          `json:"user"`
	PythonInterpreter string          `json:"ansible_python_interpreter,omitempty"`
	Repository        repository      `json:"repository"`
	Files             []customization `json:"files"`
	Containers        []docker        `json:"containers"`
}

func (a *Action) setUser(oses []*os, user string) error {
	os := &os{}
	if len(oses) == 0 {
		return errors.New(utilities.UNKOWN_OSES)
	} else if len(oses) == 1 {
		os = oses[0]
	} else {
		for _, k := range oses {
			if k.User == user {
				os = k
				break
			}
		}
		if os.User == "" {
			return errors.New(utilities.UNKOWN_OSES)
		}
	}
	a.User = os.User
	a.PythonInterpreter = os.PythonInterpreter
	return nil
}
