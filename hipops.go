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
	Branch, Config, Data, Host,
	Image, Name, Repo, Dir, Run,
	RunCustom, Server, SshKey, Type string
	Ports []int
	Cred  Cred
}

type Cred struct {
	DbName, Username, Password string
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
	var pat = regexp.MustCompile(`\$[A-Z_-]+`)
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
		log.Fatal("Help: --help")
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
				for _, action := range p.Actions {
					dockerParams := configureParams(parse(setAnsibleParams(action.Params, true), c, app), false)
					containerName := extractName(dockerParams)
					if containerName == "" {
						containerName = parse(p.Name, c, app)
						if containerName == "" {
							log.Fatalf("ERROR: container name is required.")
						}
						dockerParams = fmt.Sprintf("--name %s %s", containerName, dockerParams)
					}
					RunCmd("ansible-playbook",
						fmt.Sprintf("%s%s", *flPlaybooks, p.Play),
						"-i", *flHosts,
						"-u ubuntu",
						"--private-key", *flPrivateKey,
						"-e", fmt.Sprintf("inventory=%s name=%s image=%s state=%s params=\"%s\" repo=%s sshKey=%s branch=%s dir=%s path=%s runSource=%s runDest=%s",
							p.Inventory,
							containerName,
							parse(action.Image, c, app),
							p.State,
							dockerParams,
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
			}
		} else {
			for _, action := range p.Actions {
				dockerParams := parse(action.Params, c, "")
				containerName := extractName(dockerParams)
				if containerName == "" {
					containerName = parse(p.Name, c, "")
					if containerName == "" {
						log.Fatalf("ERROR: container name is required.")
					}
					dockerParams = fmt.Sprintf("--name %s %s", containerName, dockerParams)
				}
				RunCmd("ansible-playbook",
					fmt.Sprintf("%s%s", *flPlaybooks, p.Play),
					"-i", *flHosts,
					"-u ubuntu",
					"--private-key", *flPrivateKey,
					"-e", fmt.Sprintf("inventory=%s name=%s image=%s state=%s params=\"%s\" repo='' sshKey='' branch='' dir='' path='' runSource=NA",
						p.Inventory,
						containerName,
						parse(action.Image, c, ""),
						p.State,
						dockerParams,
					),
					*flDebug,
				)
			}
		}
	}
}

func extractName(s string) string {
	var pat = regexp.MustCompile(`--name\s.+\s`)
	match := pat.FindString(s)
	if match != "" {
		name := strings.Split(pat.FindString(s), " ")[1]
		return name
	}
	return match
}

func setAnsibleParams(s string, set bool) string {
	if set == true {
		var p = regexp.MustCompile(`({{\s*ansible_(&}})*)`)
		return p.ReplaceAllString(s, "@ANSIBLE.${2}")
	}
	var p = regexp.MustCompile(`(@ANSIBLE.(&}})*)`)
	return p.ReplaceAllString(s, "{{ ansible_${2}")
}

func configureParams(input string, set bool) string {
	input = setAnsibleParams(input, false)
	env := envMap{}
	for _, v := range os.Environ() {
		var a = strings.Split(v, "=")
		env[a[0]] = a[1]
	}

	input = expandEnv(input, env.Get)
	return input

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
