package parser

import (
	"fmt"
	"testing"

	"github.com/aminjam/hipops/plugins"
	"github.com/aminjam/hipops/utilities"
)

const scenario = `
  "id": "0",
  "description": "",
  "env": "test",
  "dest": "/data"
`
const oses = `
  ,"oses": [{
    "user": "core",
    "pythonInterpreter": "PATH=/home/core/bin:$PATH python"
  }]
`
const apps = `
  ,"apps": [{
    "name": "mongo",
    "type": "db",
    "image": "aminjam/mongodb:latest",
    "ports": [27017]
  }]
`
const playbooks = `
  ,"playbooks": [{
    "inventory": "tag_App-Role_SAMOMY-DEV",
    "apps": ["{{index .Apps 0}}"],
    "containers": [{
      "params": "-v {{.App.Dest}}:/home/app -p 9990:{{index .App.Ports 0}} -e MONGO_OPTIONS='--smallfiles' -d {{.App.Image}} /run.sh"
    }]
  }]
`

var testPlugin plugins.Plugin

type instance struct{}

func (i *instance) DefaultPlay() string        { return "test.play" }
func (i *instance) Mask(input string) string   { return input }
func (i *instance) Unmask(input string) string { return input }
func init()                                    { testPlugin = &instance{} }

func TestScenarioParse_HappyPath(t *testing.T) {
	config := []byte(fmt.Sprintf("{%s%s%s%s}", scenario, oses, apps, playbooks))
	var sc Scenario
	actions, err := sc.Parse(config, testPlugin)

	spec := utilities.Spec(t)
	spec.Expect(err).ToEqual(nil)
	spec.Expect(actions[0].User).ToEqual("core")
	spec.Expect(actions[0].Play).ToEqual("test.play")
	spec.ExpectString(actions[0].Containers[0].Params).ToContain("27017")
	spec.ExpectString(actions[0].Containers[0].Params).ToContain("--name 0-db-mongo")

}

func TestScenarioParse_UnhappyPath(t *testing.T) {

	spec := utilities.Spec(t)

	const playbooks_missing_app = `
  ,"playbooks": [{
    "inventory": "tag_App-Role_SAMOMY-DEV",
    "apps": ["{{index .Apps 1}}"],
    "containers": [{
      "params": "-v {{.App.Dest}}:/home/app -p 9990:{{index .App.Ports 0}} -e MONGO_OPTIONS='--smallfiles' -d {{.App.Image}} /run.sh"
    }]
  }]
`
	config := []byte(fmt.Sprintf("{%s%s%s%s}", scenario, oses, apps, playbooks_missing_app))
	var sc0 Scenario
	_, err := sc0.Parse(config, testPlugin)
	spec.Expect(err.Error()).ToEqual(utilities.APP_NOT_FOUND)

	config = []byte(fmt.Sprintf("{%s%s%s}", scenario, apps, playbooks_missing_app))
	var sc1 Scenario
	_, err = sc1.Parse(config, testPlugin)
	spec.Expect(err.Error()).ToEqual(utilities.UNKOWN_OSES)

	const scenario_unkown_dest = `
  "id": "0",
  "description": "",
  "env": "test"
`
	config = []byte(fmt.Sprintf("{%s%s%s%s}", scenario_unkown_dest, oses, apps, playbooks))
	var sc2 Scenario
	_, err = sc2.Parse(config, testPlugin)
	spec.Expect(err.Error()).ToEqual(utilities.UNKNOWN_SCENARIO_DEST)

	const apps_invalid_repo = `
  ,"apps": [{
    "name": "mongo",
    "type": "db",
    "image": "aminjam/mongodb:latest",
    "repository":{
      "sshUrl": "git@github.com:aminjam/beersample-node.git"
    },
    "ports": [27017]
  }]
`
	config = []byte(fmt.Sprintf("{%s%s%s%s}", scenario, oses, apps_invalid_repo, playbooks))
	var sc3 Scenario
	_, err = sc3.Parse(config, testPlugin)
	spec.Expect(err.Error()).ToEqual(utilities.INVALID_REPOSITORY)
}
