package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/docker/engine-api/client"
)

// Pipeline define the various steps to be executed to produce artifact
type Stages map[string]Stage

type Pipeline struct {
	stages  Stages
	Project string
}

// Stages return ordered slice of stages, as described in the config file.
func (p Pipeline) Stages() []Stage {
	stages := []Stage{}
	for i := 0; i < len(p.stages); i++ {
		for k, s := range p.stages {
			if s.Order == i {
				s.Name = k
				stages = append(stages, s)
			}
		}
	}
	return stages
}

// in-memory stash store
// until we plug a tmp-file based implementation for large contents
var stash = make(map[string][]byte)

// Stash some binary content so it can later be retrieved by name for further usage in pipeline
func (p *Pipeline) Stash(name string, content []byte) error {
	stash[name] = content
	return nil
}

// UnStash some stashed content identified by name
func (p *Pipeline) UnStash(name string) ([]byte, error) {
	data, ok := stash[name]
	var err error
	if !ok {
		err = errors.New("No stash named " + name)
	}
	return data, err
}

func (p Pipeline) String() string {
	st := "\n"
	for i, s := range p.Stages() {
		st = st + "#" + strconv.Itoa(i) + " :: " + s.String() + "\n"
	}
	return st
}

// Stage defines a  set of commands we run in a docker container
type Stage struct {
	Order int // injected during yml parsing
	Name  string
	Exec  Exec // the actual execution of this stage
}

// Exec defines what has to run during stage execution
type Exec interface {
	fmt.Stringer
	Run(docker *client.Client, p *Pipeline, s *Stage) error
}

func (s Stage) String() string {
	return strconv.Itoa(s.Order) + ":" + s.Name + ":" + s.Exec.String()
}
