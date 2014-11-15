package command

import (
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/aminjam/hipops/parser"
	"github.com/aminjam/hipops/plugins"
	"github.com/aminjam/hipops/plugins/ansible"
	"github.com/aminjam/hipops/utilities"
	"github.com/mitchellh/cli"
)

var myPlugins []*plugins.Plugin

type params struct {
	baseDir, config, gitKey, plugin,
	privateKey, trigger string
	debug int

	//ansible plugin
	inventory, playbookPath string
}

func (p *params) toAction(a *plugins.Action) {
	if a.Repository != nil && a.Repository.SshKey == "" {
		a.Repository.SshKey = p.gitKey
	}
	p.playbookPath = strings.TrimSuffix(p.playbookPath, "/")
	if p.playbookPath != "" {
		a.Play = fmt.Sprintf("%s/%s", p.playbookPath, a.Play)
	}
	a.InventoryFile = p.inventory
	a.PrivateKey = p.privateKey
	a.Debug = p.debug
	baseDir := filepath.Dir(p.config)
	for i, _ := range a.Files {
		if strings.Contains(a.Files[i].Src, "@BASEDIR") {
			a.Files[i].Src, _ = filepath.Abs(strings.Replace(a.Files[i].Src, "@BASEDIR", baseDir, -1))
		}
	}
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
	plugin := myPlugins[0]
	switch c.params.plugin {
	case "ansible":
		err := (*plugin).ValidateParams(c.params.inventory, c.params.playbookPath)
		utilities.CheckErr(err)
	}
	config, err := ioutil.ReadFile(c.params.config)
	utilities.CheckErr(err)

	var scenario parser.Scenario
	err = scenario.Configure(config)
	utilities.CheckErr(err)
	utilities.CleanupTempFiles(scenario.Suffix)

	actions, err := scenario.Parse(plugin)
	utilities.CheckErr(err)
	for _, a := range actions {
		if c.params.trigger == "" || a.State() == utilities.DEFAULT_APP_STATE || (c.params.trigger != "" && strings.HasSuffix(a.Name, c.params.trigger)) {
			c.params.toAction(a)
			err = (*plugin).Run(a)
			utilities.CheckErr(err)
		}
	}

	c.Ui.Info(scenario.Id)

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

func init() {
	myPlugins = []*plugins.Plugin{&ansible.Instance}
}
