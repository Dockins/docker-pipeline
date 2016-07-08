package main

import (
	"strings"

	"gopkg.in/yaml.v2"
)

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

type stageSpec struct {
	// Command
	Image    string
	Env      map[string]string
	Commands []string
	Shell    string
	Cached   []string

	// Build
	Build       string
	Dockerfile  string
	ContextPath string
}

var i = 0

// UnmarshalYAML implements yaml.v2 Unmarshaler interface to inject an Order attribute
// while the docker-pipeline yaml file is parsed, as go map isn't ordered (by design).
func (s *Stage) UnmarshalYAML(unmarshal func(interface{}) error) error {
	st := stageSpec{}
	if err := unmarshal(&st); err != nil {
		return err
	}
	if st.Image != "" {
		s.Exec = Command{
			Image:    st.Image,
			Commands: st.Commands,
			Shell:    st.Shell,
			Cached:   st.Cached,
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

	s.Order = i
	i = i + 1
	return nil
}
