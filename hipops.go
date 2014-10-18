package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	flag "github.com/dotcloud/docker/pkg/mflag"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"
)

type Scenario struct {
	Env, Id, Description, Dest string
	Oses                       []Os
	Apps                       []App
	Playbooks                  []Playbook
}

type Os struct {
	Name, User, Python_interpreter string
}

type App struct {
	Host, Image, Name,
	Dest, Type string
	Ports          []int
	Cred           Cred
	Customizations []Customization
	Repository     Repository
}

func (a *App) configureName(scenario *Scenario) {
	if a.Type != "global" {
		a.Name = fmt.Sprintf("%s-%s-%s-%s", scenario.Id, scenario.Env, a.Type, a.Name)
	}
}

type Repository struct {
	Branch string `json:"branch"`
	SshUrl string `json:"sshUrl"`
	SshKey string `json:"sshKey"`
	Folder string `json:"folder"`
}

type Customization struct {
	Src        string `json:"src"`
	Dest       string `json:"dest"`
	DestFolder string `json:"destFolder"`
	Mode       int    `json:"mode"`
}

type Cred struct {
	DbName, Username, Password string
}

type Playbook struct {
	Name, Play, State,
	Inventory, User string
	Actions []DockerAction
	Apps    []string
}

type DockerAction struct {
	Image  string `json:"image"`
	Params string `json:"params"`
	Name   string `json:"name"`
	State  string `json:"state"`
}

type AnsibleVars struct {
	Dest               string          `json:"dest"`
	Inventory          string          `json:"inventory"`
	Python_interpreter string          `json:"ansible_python_interpreter,omitempty"`
	Repository         Repository      `json:"repository"`
	Files              []Customization `json:"files"`
	Containers         []DockerAction  `json:"containers"`
}

func (av *AnsibleVars) parseActions(scenario *Scenario, p *Playbook, appString string) {
	for _, action := range p.Actions {
		params := configureParams(parse(setAnsibleParams(action.Params, true), *scenario, appString), false)
		name := extractName(params)
		if name == "" {
			name = parse(p.Name, *scenario, appString)
			if name == "" {
				err := errors.New(fmt.Sprintf("Error, container name is required for play %s", p.Name))
				check(err)
			}
			params = fmt.Sprintf("--name %s %s", name, params)
		}
		if action.Image == "" {
			action.Image = "{{.App.Image}}"
		}
		action.Image = parse(action.Image, *scenario, appString)
		av.Containers = append(av.Containers, DockerAction{Name: name, Image: action.Image, Params: params, State: p.State})
	}
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
		flTrigger    = flag.String([]string{"t", "-trigger"}, "", "name of the app to trigger")
	)
	flag.Parse()
	if *flHosts == "" {
		log.Fatal("Help: --help")
	}
	if *flDebug != "" {
		*flDebug = "-" + strings.TrimPrefix(*flDebug, "-")
	}

	config, err := ioutil.ReadFile(*flConfigFile)
	configDir := filepath.Dir(*flConfigFile)
	playbooksPath := strings.TrimSuffix(*flPlaybooks, "/") + "/"

	check(err)
	var scenario Scenario
	err = json.Unmarshal(config, &scenario)
	check(err)
	fileSuffix := fmt.Sprintf("%s-%s", scenario.Id, scenario.Env)

	scenario.Dest = strings.TrimSuffix(scenario.Dest, "/")
	cleanupFiles(fileSuffix)
	for i, _ := range scenario.Apps {
		scenario.Apps[i].configureName(&scenario)
	}
	for _, p := range scenario.Playbooks {
		ansibleVars := AnsibleVars{}
		repo := Repository{SshKey: *flGitKey}
		ansibleVars.Repository = repo

		os := Os{}
		if len(scenario.Oses) == 0 {
			check(errors.New("oses is unkown."))
		} else if len(scenario.Oses) == 1 {
			if p.User != "" && p.User != scenario.Oses[0].User {
				os.User = p.User
			} else {
				os = scenario.Oses[0]
			}
		} else {
			if p.User == "" {
				check(errors.New("multiple Oses are defined, but playbook lack the user."))
			}
			for _, k := range scenario.Oses {
				if k.User == p.User {
					os = k
					break
				}
			}
			if os.User == "" {
				os.User = p.User
			}
		}
		ansibleVars.Python_interpreter = os.Python_interpreter

		if p.State == "" {
			p.State = "running"
		}

		if len(p.Apps) != 0 {
			for _, appString := range p.Apps {
				p.Name = parse("{{.App.Name}}", scenario, appString)
				if *flTrigger == "" || p.State == "running" || (*flTrigger != "" && strings.HasSuffix(p.Name, *flTrigger)) {
					err, app := findapp(p.Name, scenario)
					if err != nil {
						err := errors.New(fmt.Sprintf("Error, %s: %s", err, appString))
						check(err)
					} else {
						fmt.Println(fmt.Sprintf("Running: app %s:%s", appString, app.Name))
					}
					if p.Play == "" {
						p.Play = "hipops.yml"
					}

					if app.Dest == "" {
						if scenario.Dest == "" {
							err := errors.New("scenario and app-specific destination paths are both unkown.")
							check(err)
						}
						if app.Type == "" {
							app.Type = "generic"
						}
						app.Dest = fmt.Sprintf("%s/%s-%s/%s/%s", scenario.Dest, scenario.Id, scenario.Env, app.Type, app.Name)
					}
					ansibleVars.Dest = app.Dest

					if app.Customizations != nil {
						for _, customization := range app.Customizations {
							src := customization.Src
							dest := customization.Dest
							if strings.HasPrefix(src, "http") {
								src = downloadFile(src, fileSuffix)
							} else if !strings.HasPrefix(src, "/") {
								src, _ = filepath.Abs(configDir + "/" + src)
							}
							mode := customization.Mode
							if mode == 0 {
								mode = 400
							}
							if !strings.HasPrefix(dest, "~") || !strings.HasPrefix(dest, "/") {
								dest = fmt.Sprintf("%s/%s", strings.TrimSuffix(ansibleVars.Dest, "/"), dest)
							}
							destFolder := dest[:strings.LastIndex(dest, "/")]
							ansibleVars.Files = append(ansibleVars.Files, Customization{Src: src, DestFolder: destFolder, Dest: dest, Mode: mode})
						}
					} else {
						ansibleVars.Files = make([]Customization, 0)
					}

					if app.Repository.SshUrl != "" {
						if ansibleVars.Repository.SshKey != "" {
							app.Repository.SshKey = ansibleVars.Repository.SshKey
						}
						if app.Repository.SshKey == "" {
							err := errors.New(fmt.Sprintf("app %s has no associated repository SSH-Key", app.Name))
							check(err)
						}
						if app.Repository.Branch == "" {
							app.Repository.Branch = "master"
						}

						ansibleVars.Repository.SshKey = app.Repository.SshKey
						ansibleVars.Repository.SshUrl = app.Repository.SshUrl
						ansibleVars.Repository.Branch = app.Repository.Branch
						ansibleVars.Repository.Folder = app.Repository.Folder
					}

					ansibleVars.Inventory = p.Inventory
					ansibleVars.parseActions(&scenario, &p, appString)
					content, err := json.Marshal(ansibleVars)
					check(err)
					fileName := writeFile(content, "json", fileSuffix)
					RunCmd("ansible-playbook",
						fmt.Sprintf("%s%s", playbooksPath, p.Play),
						"-i", *flHosts,
						"-u", os.User,
						"--private-key", *flPrivateKey,
						"--extra-vars", "@"+fileName,
						*flDebug,
					)
				}
			}
		} else {
			if p.Play == "" {
				p.Play = "hipops.yml"
			}
			ansibleVars.Inventory = p.Inventory
			ansibleVars.Files = make([]Customization, 0)
			ansibleVars.parseActions(&scenario, &p, "")
			content, err := json.Marshal(ansibleVars)
			check(err)
			fileName := writeFile(content, "json", fileSuffix)
			RunCmd("ansible-playbook",
				fmt.Sprintf("%s%s", playbooksPath, p.Play),
				"-i", *flHosts,
				"-u", os.User,
				"--private-key", *flPrivateKey,
				"--extra-vars", "@"+fileName,
				*flDebug,
			)
		}
	}
}

func downloadFile(url string, suffix string) string {
	rand.Seed(time.Now().UnixNano())
	fileName := fmt.Sprintf("/tmp/hipops-%s-%v", suffix, rand.Intn(1000000))
	fmt.Println("Downloading file...")

	output, err := os.Create(fileName)
	defer output.Close()

	response, err := http.Get(url)
	check(err)
	defer response.Body.Close()

	_, err = io.Copy(output, response.Body)
	check(err)
	return fileName
}

func cleanupFiles(suffix string) {
	d, err := os.Open("/tmp")
	defer d.Close()
	check(err)

	files, err := d.Readdir(-1)
	check(err)

	fmt.Println("Reading files for /tmp")

	for _, file := range files {
		if file.Mode().IsRegular() {
			if strings.HasPrefix(file.Name(), "hipops-"+suffix) {
				os.Remove("/tmp/" + file.Name())
				fmt.Println("Deleted ", file.Name())
			}
		}
	}
}

func writeFile(content []byte, fileType string, suffix string) string {
	rand.Seed(time.Now().UnixNano())
	fileName := fmt.Sprintf("/tmp/hipops-%s-%v.%s", suffix, rand.Intn(1000000), fileType)
	output, err := os.Create(fileName)
	defer output.Close()

	check(err)
	_, err = io.WriteString(output, fmt.Sprintf("%s", content))
	check(err)
	return fileName
}

func findapp(name string, scenario Scenario) (error, *App) {
	if name != "" {
		for k, v := range scenario.Apps {
			if strings.Contains(v.Name, name) {
				return nil, &scenario.Apps[k]
			}
		}
	}
	return errors.New("app not found."), nil
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
