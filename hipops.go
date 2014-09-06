package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	flag "github.com/dotcloud/docker/pkg/mflag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"
	"path/filepath"
)

type Configuration struct {
	Apps      []App
	Env       string
	Id        string
	Playbooks []Playbook
	Servers   []Server
}
type App struct {
	Branch string
	Config string
	Data   string
	Host   string
	Image  string
	Name   string
	Repo   string
	Ports  []int
	Dir   string
	Run  	 string
	RunCustom string
	Server string
	SshKey string
	Type   string
}

type Server struct {
	Apps []string
	Role string
	Type string
}
type Playbook struct {
	Actions   []DockerAction
	Apps      []string
	Inventory string
	Name      string
	Play      string
	State     string
}
type DockerAction struct {
	Image  string
	Params string
}

func main() {
	var (
		flHosts      = flag.String([]string{"h", "-hosts"}, "./hosts/local", "Inventory Hosts Target e.g. local,aws")
		flConfigFile = flag.String([]string{"c", "-config"}, "./config.json", ".json configuration")
		flPlaybooks  = flag.String([]string{"p", "-playbook-path"}, "../../ansible-playbooks/", "Playbooks Path")
		flPrivateKey = flag.String([]string{"k", "-private-key"}, "", "SSH Private Key")
	)
	flag.Parse()
	if *flHosts == "" {
		log.Fatal("Usage: [-h <Inventory Hosts Target>][-k <SSH private key>]")
	}
	config, err := ioutil.ReadFile(*flConfigFile)
	check(err)
	var c Configuration
	err = json.Unmarshal(config, &c)
	check(err)
	for k, v := range c.Apps {
		c.Apps[k].Name = parse(v.Name + "-{{.Id}}-{{.Env}}", c, "")
		c.Apps[k].Data = parse(v.Data + v.Name, c, "")
	}
	for _, p := range c.Playbooks {
			if len(p.Apps) != 0 {
				for _,app := range p.Apps {
					fmt.Println(app);
					runSource, runDest := "NA","NA"
					if (parse("{{.App.RunCustom}}", c, app) != ""){
						dir := filepath.Dir(*flConfigFile)
						runSource,_ =  filepath.Abs(dir + "/" + parse("{{.App.RunCustom}}", c, app))
						runDest = strings.SplitAfter(parse("{{.App.Run}}", c, app)," ")[0]
					}
					//fmt.Println("PORT:" + parse(" -v {{.App.Data}}:/home/app -p 8002:{{index ((index .Apps 0).Ports) 0}} -d {{.App.Image}}", c, app));
					//fmt.Println("PORT:" + parse(" -v {{.App.Data}}:/home/app -p 8002:{{index .App.Ports 1}} -d {{.App.Image}}", c, app));
					RunCmd("ansible-playbook",
						fmt.Sprintf("%s%s", *flPlaybooks, p.Play),
						"-i", *flHosts,
						"-u ubuntu",
						"--private-key", *flPrivateKey,
						"-e", fmt.Sprintf("inventory=%s name=%s image=%s state=%s params=\"%s\" repo=%s sshKey=%s branch=%s dir=%s path=%s runSource=%s runDest=%s",
							p.Inventory,
							parse(p.Name, c, app),
							parse(p.Actions[0].Image, c, app),
							p.State,
							parse(p.Actions[0].Params, c, app),
							parse("{{.App.Repo}}", c, app),
							parse("{{.App.SshKey}}", c, app),
							parse("{{.App.Branch}}", c, app),
							parse("{{.App.Dir}}",c,app),
							parse("{{.App.Data}}", c, app),
							runSource,runDest,
						),
						"-vvvv",
					)
				}
			} else {
				RunCmd("ansible-playbook",
					fmt.Sprintf("%s%s", *flPlaybooks, p.Play),
					"-i", *flHosts,
					"-u ubuntu",
					"--private-key", *flPrivateKey,
					"-e", fmt.Sprintf("inventory=%s name=%s image=%s state=%s params=\"%s\"",
						p.Inventory,
						parse(p.Name, c, ""),
						parse(p.Actions[0].Image, c, ""),
						p.State,
						parse(p.Actions[0].Params, c, ""),
					),
					"-vvvv",
				)
			}
	}
}
func format(input string, app string) string {
	app = strings.Replace(app, "{{", "(", -1)
	app = strings.Replace(app, "}}", ")", -1)
	var re = regexp.MustCompile(`({{.App(.&}})*)`)
	input = re.ReplaceAllString(input, fmt.Sprintf("{{%s${2}", app))
	re = regexp.MustCompile(`({{index .App.Ports(.&}})*)`)
	input = re.ReplaceAllString(input, fmt.Sprintf("{{index (%s.Ports)${2}", app))
	return input
}
func parse(input string, base interface{}, app string) string {
	t := template.New("")
	if app != "" {
		input = format(input, app)
	}
	t, _ = t.Parse(input)
	buf := new(bytes.Buffer)
	t.Execute(buf, base)
	return buf.String()
}

func check(err error) {
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}
func RunCmd(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	stdout, err := cmd.StdoutPipe()
	check(err)
	stderr, err := cmd.StderrPipe()
	check(err)
	err = cmd.Start()
	check(err)
	defer cmd.Wait()
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
}
