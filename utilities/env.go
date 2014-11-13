package utilities

import (
	"os"
	"regexp"
	"strings"
)

type environmentMap func(string) string
type envMap map[string]string

func (s *envMap) Get(n string) string {
	n = strings.Replace(n, "$", "", 1)
	return (*s)[n]
}

func expandEnv(s string, m environmentMap) string {
	var pat = regexp.MustCompile(`\$[A-Z_-]+`)
	return string(pat.ReplaceAllFunc([]byte(s), func(bs []byte) []byte {
		return []byte(m(string(bs)))
	}))
}

func ParseEnvFlags(input string) string {
	env := envMap{}
	for _, v := range os.Environ() {
		var a = strings.Split(v, "=")
		env[a[0]] = a[1]
	}

	input = expandEnv(input, env.Get)
	return input
}
