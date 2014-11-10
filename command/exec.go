package command

import (
	"flag"
	"github.com/aminjam/hipops/parser"
	"github.com/aminjam/hipops/utilities"
	"github.com/mitchellh/cli"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type params struct {
	config, gitKey, plugin,
	privateKey, trigger string
	debug int

	//ansible plugin
	inventory, playbookPath string
}

type ExecCommand struct {
	ShutdownCh <-chan struct{}
	Ui         cli.Ui
	params     params
	sessionID  string
}

func (c *ExecCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("exec", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }
	//common flags
	cmdFlags.StringVar(&c.params.config, "config", "./config.json", "")
	cmdFlags.IntVar(&c.params.debug, "debug", 0, "")
	cmdFlags.StringVar(&c.params.gitKey, "git-key", "~/.ssh/id_rsa", "")
	cmdFlags.StringVar(&c.params.plugin, "plugin", "", "")
	cmdFlags.StringVar(&c.params.privateKey, "private-key", "", "")
	cmdFlags.StringVar(&c.params.trigger, "trigger", "", "")

	//ansible plugin flags
	cmdFlags.StringVar(&c.params.inventory, "inventory", "./hosts/local", "")
	cmdFlags.StringVar(&c.params.playbookPath, "playbook-path", "", "")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	// Check for a plugin
	if c.params.plugin == "" {
		c.Ui.Error(utilities.UNKOWN_PLUGIN)
		c.Ui.Error("--------")
		c.Ui.Error(c.Help())
		return 1
	}
	config, err := ioutil.ReadFile(c.params.config)
	baseDir := filepath.Dir(c.params.config)
	utilities.CheckErr(err)

	var scenario parser.Scenario
	_, err = scenario.Parse(config)
	utilities.CheckErr(err)

	c.Ui.Info(scenario.Id)
	c.Ui.Info(baseDir)

	utilities.CleanupTempFiles(scenario.Suffix)

	return 0
}

func (c *ExecCommand) Synopsis() string {
	return "Executes a JSON scenerio with a plugin"
}
func (c *ExecCommand) Help() string {
	helpText := `
Usage: hipops exec [options] [-|command...]
Executes a JSON scenerio with a plugin
Options:
	-config="./config.json"    hipops JSON configuration
	-debug=0                   debug level (0-3)
	-git-key="~/.ssh/id_rsa"   SSH Git Key for Repo
	-plugin=""                 Name of the plugin (e.g. ansible)
	-private-key=""            SSH Host Private Key
	-trigger=""                Name of the app to trigger

	(ansible plugin)
	-inventory="./hosts/local"     Inventory Hosts Target
	-playbook-path=""              Ansible Playbook Path (Optional)
`
	return strings.TrimSpace(helpText)
}
