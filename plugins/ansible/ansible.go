package plugins

import (
	"github.com/aminjam/hipops/parser"
)

var Ansible parser.Plugin

type instance struct{}

func init() {
	Ansible = &instance{}
}
func (i *instance) Mask() error {
	return nil
}
