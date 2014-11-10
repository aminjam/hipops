package utilities

import (
	"strings"
	"testing"
)

type S struct {
	t *testing.T
}

type SR struct {
	t        *testing.T
	expected []interface{}
	str      string
}

func Spec(t *testing.T) *S {
	return &S{t: t}
}

func (s *S) Expect(expected ...interface{}) (sr *SR) {
	return &SR{t: s.t, expected: expected}
}
func (s *S) ExpectString(str string) (sr *SR) {
	return &SR{t: s.t, str: str}
}

func (sr *SR) ToEqual(actuals ...interface{}) {
	for index, actual := range actuals {
		if sr.expected[index] != actual {
			sr.t.Errorf("expected %+v to equal %+v", sr.expected[index], actual)
		}
	}
}

func (sr *SR) ToNotEqual(actuals ...interface{}) {
	for index, actual := range actuals {
		if sr.expected[index] == actual {
			sr.t.Errorf("expected %+v to not equal %+v", sr.expected[index], actual)
		}
	}
}

func (sr *SR) ToContain(actual string) {
	if !strings.Contains(sr.str, actual) {
		sr.t.Errorf("expected %+v to contain %+v", sr.str, actual)
	}
}
