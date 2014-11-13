package ansible

import (
	"regexp"

	"github.com/aminjam/hipops/plugins"
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
