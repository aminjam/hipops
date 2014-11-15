package ansible

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"regexp"

	"github.com/aminjam/hipops/plugins"
	"github.com/aminjam/hipops/utilities"
)

var Instance plugins.Plugin

type instance struct{}

func init() {
	Instance = &instance{}
}
func (i *instance) DefaultPlay() string {
	return "hipops.yml"
}
func (i *instance) Mask(input string) string {
	var p = regexp.MustCompile(`({{\s*ansible_(&}})*)`)
	return p.ReplaceAllString(input, "@ANSIBLE.${2}")
}
func (i *instance) Unmask(input string) string {
	var p = regexp.MustCompile(`(@ANSIBLE.(&}})*)`)
	return p.ReplaceAllString(input, "{{ ansible_${2}")
}
func (i *instance) Run(a *plugins.Action) error {
	content, err := json.Marshal(a)
	if err != nil {
		return err
	}
	fileName, err := utilities.WriteFile(content, "json", a.Suffix)
	if err != nil {
		return err
	}
	params := []string{
		a.Play,
		"-i", a.InventoryFile,
		"-u", a.User,
		"--private-key", a.PrivateKey,
		"--extra-vars", "@" + fileName,
	}
	switch a.Debug {
	case 1:
		params = append(params, "-v")
	case 2:
		params = append(params, "-vv")
	case 3:
		params = append(params, "-vvv")
	}
	return utilities.RunCmd("ansible-playbook", params...)
}
func (i *instance) ValidateParams(args ...string) error {
	var inventoryFile = args[0]
	var playbookPath = args[1]
	if _, err := filepath.Abs(inventoryFile); err != nil {
		return err
	}
	if playbookPath == "" {
		return errors.New("--playbook-path is required for ansible plugin")
	}
	if _, err := filepath.Abs(playbookPath); err != nil {
		return err
	}
	return nil
}
