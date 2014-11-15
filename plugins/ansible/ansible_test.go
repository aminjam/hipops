package ansible

import (
	"strings"
	"testing"

	"github.com/aminjam/hipops/plugins"
	"github.com/aminjam/hipops/utilities"
)

type tuple struct{ org, masked string }

func TestAnsiblePlugin_implements(t *testing.T) {
	var _ plugins.Plugin = &instance{}
}
func TestAnsiblePlugin_masking(t *testing.T) {
	spec := utilities.Spec(t)
	i := &instance{}

	test_values := []tuple{
		{"{{ box_hostname }}", "@ANSIBLE.hostname }}"},
		{"--name mine -h {{ box_hostname }} -d myimage", "--name mine -h @ANSIBLE.hostname }} -d myimage"},
		{"--name mine -h {{ box_hostname }} -p {{ box_ip}}:80 -d myimage", "--name mine -h @ANSIBLE.hostname }} -p @ANSIBLE.ip}}:80 -d myimage"},
	}
	for _, entry := range test_values {
		val := i.Mask(entry.org)
		spec.ExpectString(val).ToContain(entry.masked)
	}
	for _, entry := range test_values {
		val := i.Unmask(entry.masked)
		spec.ExpectString(val).ToContain(strings.Replace(entry.org, "box", "ansible", -1))
	}
}
