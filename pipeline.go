package main

import (
	"strings"

	"github.com/docker/engine-api/client"
	"gopkg.in/yaml.v2"
)

// Pipeline define the various steps to be executed to produce artifact
type Pipeline map[string]Stage

// Parse yml data to produce a Pipeline data structure
func Parse(source []byte) (Pipeline, error) {
	var pipeline Pipeline
	err := yaml.Unmarshal(source, &pipeline)
	if err != nil {
		return nil, err
	}

	return pipeline, nil
}

// UnmarshalYAML implements yaml.v2 Unmarshaler interface to set Stage.Name reflecting map's key
func (p *Pipeline) UnmarshalYAML(unmarshal func(interface{}) error) error {
	ps := map[string]Stage{}
	if err := unmarshal(&ps); err != nil {
		return err
	}
	for k, s := range ps {
		s.Name = k
	}
	*p = ps
	return nil
}

// Stage defines a  set of commands we run in a docker container
type Stage struct {
	Order int // injected during yml parsing
	Name  string
	Exec  Exec // the actual execution of this stage
}

type stageSpec struct {
	// Command
	Image    string
	Env      map[string]string
	Commands []string

	// Build
	Build       string
	Dockerfile  string
	ContextPath string

	// Push
	Push		string
	Registry 	string
}

// Exec defines what has to run during stage execution
type Exec interface {
	Run(docker *client.Client, s Stage) error
}

var i = 0

// UnmarshalYAML implements yaml.v2 Unmarshaler interface to inject an Order attribute
// while the docker-pipeline yaml file is parsed, as go map isn't ordered (by design).
func (s *Stage) UnmarshalYAML(unmarshal func(interface{}) error) error {
	st := stageSpec{}
	if err := unmarshal(&st); err != nil {
		return err
	}
	s.Order = i
	if st.Image != "" {
		s.Exec = Command{
			Image:    st.Image,
			Commands: st.Commands,
			Env:      st.Env,
		}
	}
	if st.Build != "" {
		s.Exec = Build{
			Tags:        strings.Split(st.Build, " "),
			Dockerfile:  st.Dockerfile,
			ContextPath: st.ContextPath,
		}
	}
	if st.Push != "" {
		s.Exec = Push{
			Tags:        strings.Split(st.Build, " "),
			Registry: st.Registry
		}
	}
	
	i = i + 1
	return nil
}
