package parser

import (
	"errors"

	"github.com/aminjam/hipops/utilities"
)

type Action struct {
	Dest              string           `json:"dest"`
	Inventory         string           `json:"inventory"`
	User              string           `json:"user"`
	Play              string           `json:"play"`
	PythonInterpreter string           `json:"ansible_python_interpreter,omitempty"`
	Repository        *repository      `json:"repository"`
	Files             []*customization `json:"files"`
	Containers        []*container     `json:"containers"`
}

func (a *Action) fromApp(app *app) {
	a.Dest = app.Dest
	a.Repository = app.Repository
	a.Files = app.Customizations
}
func (a *Action) fromPlaybook(p *playbook) {
	a.Play = p.Play
	a.Containers = p.Containers
}
func (a *Action) fromOS(oses []*os, user string) error {
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
