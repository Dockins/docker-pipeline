package main

import (
	"fmt"
	"github.com/docker/engine-api/client"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// Pipeline define the various steps to be executed to produce artifact
type Pipeline map[string]Stage

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
	Order int    // injected during yml parsing
	Name  string 
	Exec  Exec   // the actual execution of this stage
}

type stageSpec struct {
	Image    string
	Env      map[string]string
	Commands []string
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
		s.Exec = &Command{
			Image:    st.Image,
			Commands: st.Commands,
			Env:      st.Env,
		}
	}
	i = i + 1
	return nil
}

func main() {

	f := "docker-pipeline.yml" // TODO cli flag -f

	source, err := ioutil.ReadFile(f)
	if err != nil {
		panic(err)
	}

	var pipeline Pipeline
	err = yaml.Unmarshal(source, &pipeline)
	if err != nil {
		panic(err)
	}

	docker, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(pipeline); i++ {
		for k, s := range pipeline {
			if s.Order == i {
				s.Name = k
				runStage(docker, s)
			}
		}
	}

}

func runStage(docker *client.Client, s Stage) error {
	fmt.Printf("-----------------------------------------\n")
	fmt.Printf(" Stage: %s\n", s.Name)
	fmt.Printf("-----------------------------------------\n")

	return s.Exec.Run(docker, s)
}


