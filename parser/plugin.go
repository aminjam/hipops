package parser

import "github.com/aminjam/hipops/plugins/ansible"

type Plugin interface {
	Mask() error
}

var Plugins []Plugin

func init() {
	Plugins = []Plugin{plugins.Ansible}
}
