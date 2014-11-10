package parser

import (
	"fmt"
	"github.com/aminjam/hipops/utilities"
	"testing"
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
    "actions": [{
      "params": "-v {{.App.Dest}}:/home/app -p 9990:{{index .App.Ports 0}} -e MONGO_OPTIONS='--smallfiles' -d {{.App.Image}} /run.sh"
    }]
  }]
`

func TestScenarioParse(t *testing.T) {
	config := []byte(fmt.Sprintf("{%s%s%s%s}", scenario, oses, apps, playbooks))
	var scenario Scenario
	actions, err := scenario.Parse(config)

	spec := utilities.Spec(t)
	spec.Expect(err).ToEqual(nil)
	spec.Expect(actions[0].User).ToEqual("core")

}
