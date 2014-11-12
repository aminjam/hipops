package ansible

import "github.com/aminjam/hipops/plugins"

var Instance plugins.Plugin

type instance struct{}

func init() {
	Instance = &instance{}
}
func (i *instance) Mask() error {
	return nil
}
