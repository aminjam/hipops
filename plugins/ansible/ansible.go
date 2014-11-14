package ansible

import (
	"encoding/json"
	"fmt"
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
func (i *instance) Run(action *plugins.Action, config *plugins.PluginConfig) error {
	if config.PlaybookPath == "" {
		config.PlaybookPath = "./playbooks"
	}
	content, err := json.Marshal(action)
	if err != nil {
		return err
	}
	fileName := utilities.WriteFile(content, "json", action.Name)
	params := []string{
		fmt.Sprintf("%s/%s", config.PlaybookPath, action.Play),
		"-i", action.Inventory,
		"-u", action.User,
		"--private-key", config.PrivateKey,
		"--extra-vars", "@" + fileName,
	}
	fmt.Println(params)
	return nil
}
