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
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

type Configuration struct {
	Env, Id, Description string
	Apps                 []App
	Playbooks            []Playbook
	Servers              []Server
}
type App struct {
	Branch, Config, Data, DbName,
	Host, Image, Name,
	Repo, Dir, Run, RunCustom,
	Server, SshKey, Type string
	Ports []int
}

type Server struct {
	Role, Type string
	Apps       []string
}
type Playbook struct {
	Name, Play, State,
	Inventory string
	Actions []DockerAction
	Apps    []string
}
type DockerAction struct {
	Image  string
	Params string
}

type EnvironmentMap func(string) string
type envMap map[string]string

func (s *envMap) Get(n string) string {
	n = strings.Replace(n, "$", "", 1)
	return (*s)[n]
}

func expandEnv(s string, m EnvironmentMap) string {
	var pat = regexp.MustCompile(`\$[A-Z]+`)
	return string(pat.ReplaceAllFunc([]byte(s), func(bs []byte) []byte {
		return []byte(m(string(bs)))
	}))
}

func main() {
	var (
		flHosts      = flag.String([]string{"h", "-hosts"}, "./hosts/local", "Inventory Hosts Target e.g. local,aws")
		flConfigFile = flag.String([]string{"c", "-config"}, "./config.json", ".json configuration")
		flPlaybooks  = flag.String([]string{"p", "-playbook-path"}, "../../ansible-playbooks/", "Playbooks Path")
		flPrivateKey = flag.String([]string{"k", "-host-key"}, "", "SSH Host Private Key")
		flGitKey     = flag.String([]string{"g", "-git-key"}, "", "SSH Git Key for Repo")
		flDebug      = flag.String([]string{"d", "-debug"}, "v", "debug flag e.g. vvvv")
	)
	flag.Parse()
	if *flHosts == "" {
		log.Fatal("Usage: [-h <Inventory Hosts Target>][-k <SSH private key>]")
	}
	if *flDebug != "" {
		*flDebug = "-" + strings.TrimPrefix(*flDebug, "-")
	}

	config, err := ioutil.ReadFile(*flConfigFile)
	check(err)
	var c Configuration
	err = json.Unmarshal(config, &c)
	check(err)
	for k, v := range c.Apps {
		c.Apps[k].Name = parse(v.Name+"-{{.Id}}-{{.Env}}", c, "")
		c.Apps[k].Data = parse(v.Data+v.Name, c, "")
	}
	for _, p := range c.Playbooks {
		if len(p.Apps) != 0 {
			for _, app := range p.Apps {
				fmt.Println(app)
				runSource, runDest := "NA", "NA"
				if parse("{{.App.RunCustom}}", c, app) != "" {
					dir := filepath.Dir(*flConfigFile)
					runSource, _ = filepath.Abs(dir + "/" + parse("{{.App.RunCustom}}", c, app))
					runDest = strings.SplitAfter(parse("{{.App.Run}}", c, app), " ")[0]
				}
				appGitKey := *flGitKey
				if appGitKey == "" {
					appGitKey = parse("{{.App.SshKey}}", c, app)
				}
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
						configureParams(parse(configureParams(p.Actions[0].Params, true), c, app), false),
						parse("{{.App.Repo}}", c, app),
						appGitKey,
						parse("{{.App.Branch}}", c, app),
						parse("{{.App.Dir}}", c, app),
						parse("{{.App.Data}}", c, app),
						runSource, runDest,
					),
					*flDebug,
				)
			}
		} else {
			RunCmd("ansible-playbook",
				fmt.Sprintf("%s%s", *flPlaybooks, p.Play),
				"-i", *flHosts,
				"-u ubuntu",
				"--private-key", *flPrivateKey,
				"-e", fmt.Sprintf("inventory=%s name=%s image=%s state=%s params=\"%s\" repo='' sshKey='' branch='' dir='' path='' runSource=NA",
					p.Inventory,
					parse(p.Name, c, ""),
					parse(p.Actions[0].Image, c, ""),
					p.State,
					parse(p.Actions[0].Params, c, ""),
				),
				*flDebug,
			)
		}
	}
}
func configureParams(input string, set bool) string {
	ansible_ip := "{{ ansible_eth1.ipv4.address }}"
	ansible_ip_TEMP := "@ANSIBLE_IP"
	temp := ""
	if set == true {
		temp = strings.Replace(input, ansible_ip, ansible_ip_TEMP, -1)
		return temp
	}
	temp = strings.Replace(input, ansible_ip_TEMP, ansible_ip, -1)
	env := envMap{}
	for _, v := range os.Environ() {
		var a = strings.Split(v, "=")
		env[a[0]] = a[1]
	}

	temp = expandEnv(temp, env.Get)
	return temp

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
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
	err = cmd.Wait()
	check(err)
}
