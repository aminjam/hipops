package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/aminjam/hipops/plugins"
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
	Customizations []*plugins.Customization
	Repository     *plugins.Repository
}

func (a *app) toAction(action *plugins.Action) {
	action.Dest = a.Dest
	action.Repository = a.Repository
	action.Files = a.Customizations
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
		if err := a.Customizations[c].Configure(sc.Suffix, a.Dest); err != nil {
			return err
		}
	}
	if a.Repository != nil {
		if err := a.Repository.Configure(); err != nil {
			return err
		}
	}
	return nil
}

type cred struct {
	DbName, Username, Password string
}

type playbook struct {
	Name, Play, State,
	Inventory, User string
	Containers []*plugins.Container
	Apps       []string
}

func (p *playbook) configure(plugin *plugins.Plugin) error {
	if p.Inventory == "" {
		return errors.New(utilities.INVENTORY_MISSING)
	}
	if len(p.Containers) == 0 {
		return errors.New(utilities.UNKNOWN_CONTAINERS)
	}
	if p.State == "" {
		p.State = utilities.DEFAULT_APP_STATE
	}
	if p.Play == "" {
		p.Play = (*plugin).DefaultPlay()
	}
	return nil
}

func (p *playbook) toAction(a *plugins.Action) {
	a.Name = p.Name
	a.Play = p.Play
	a.Inventory = p.Inventory
	a.Containers = p.Containers
}

type Scenario struct {
	Env, Id, Description,
	Dest, Suffix string
	Oses      []*os
	Apps      []*app
	Playbooks []*playbook
}

func (sc *Scenario) Configure(config []byte) error {
	err := json.Unmarshal(config, sc)
	if err != nil {
		return err
	}
	sc.Suffix = fmt.Sprintf("%s-%s", sc.Id, sc.Env)
	if sc.Dest == "" {
		return errors.New(utilities.UNKNOWN_SCENARIO_DEST)
	}
	sc.Dest = strings.TrimSuffix(sc.Dest, "/")

	for i, _ := range sc.Apps {
		if err = sc.Apps[i].configure(sc); err != nil {
			return err
		}
	}
	return nil
}

func (sc *Scenario) Parse(plugin *plugins.Plugin) ([]*plugins.Action, error) {
	actions, counter := make([]*plugins.Action, sc.countContainers()), 0

	for _, p := range sc.Playbooks {
		action := &plugins.Action{}
		action.Suffix = sc.Suffix
		os := &os{}
		if len(sc.Oses) == 0 {
			return nil, errors.New(utilities.UNKOWN_OSES)
		} else if len(sc.Oses) == 1 {
			os = sc.Oses[0]
		} else {
			for _, k := range sc.Oses {
				if k.User == p.User {
					os = k
					break
				}
			}
			if os.User == "" {
				return nil, errors.New(utilities.UNKOWN_OSES)
			}
		}
		action.User = os.User
		action.PythonInterpreter = os.PythonInterpreter

		p.configure(plugin)

		if len(p.Apps) != 0 {
			for _, appString := range p.Apps {
				p.Name = utilities.ParseTemplate("{{.App.Name}}", sc, appString)
				app, err := sc.findApp(p.Name)
				if err != nil {
					return nil, err
				}
				sc.configureContainers(p, plugin, appString)
				app.toAction(action)
				p.toAction(action)
			}
		} else {
			sc.configureContainers(p, plugin, "")
			p.toAction(action)
		}
		actions[counter] = action
		counter++
	}
	return actions, nil
}

func (sc *Scenario) configureContainers(p *playbook, plugin *plugins.Plugin, appString string) {
	for i, _ := range p.Containers {
		p.Containers[i].Name = p.Name
		if p.Containers[i].State == "" {
			p.Containers[i].State = p.State
		}
		masked := (*plugin).Mask(p.Containers[i].Params)
		parsed := utilities.ParseTemplate(masked, sc, appString)
		unmask := (*plugin).Unmask(parsed)
		p.Containers[i].Params = utilities.ParseEnvFlags(unmask)
		p.Containers[i].Configure()
	}
}
func (sc *Scenario) findApp(name string) (*app, error) {
	if name != "" {
		for k, v := range sc.Apps {
			if strings.Contains(v.Name, name) {
				return sc.Apps[k], nil
			}
		}
	}
	return nil, errors.New(utilities.APP_NOT_FOUND)
}
func (sc *Scenario) countContainers() int {
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
