package main

import (
	"fmt"
	"strconv"

	"github.com/docker/engine-api/client"
)

// Pipeline define the various steps to be executed to produce artifact
type Pipeline map[string]Stage

// Stages return ordered slice of stages, as described in the config file.
func (p Pipeline) Stages() []Stage {
	stages := []Stage{}
	for i := 0; i < len(p); i++ {
		for k, s := range p {
			if s.Order == i {
				s.Name = k
				stages = append(stages, s)
			}
		}
	}
	return stages
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
	Run(docker *client.Client, s Stage) error
}

func (s Stage) String() string {
	return strconv.Itoa(s.Order) + ":" + s.Name + ":" + s.Exec.String()
}
