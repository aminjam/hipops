package command

import (
	"github.com/aminjam/hipops/utilities"
	"github.com/mitchellh/cli"
	"testing"
)

func TestExecCommand_implements(t *testing.T) {
	var _ cli.Command = &ExecCommand{}
}

func TestExecCommandRun(t *testing.T) {
	ui := new(cli.MockUi)
	c := &ExecCommand{Ui: ui}
	args := []string{}

	code := c.Run(args)
	spec := utilities.Spec(t)
	if code != 0 {
		spec.ExpectString(ui.ErrorWriter.String()).ToContain(utilities.UNKOWN_PLUGIN)
	}

}
