package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/filters"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"time"
)

/* FIXME
would like to use :
```
stagename:
  image: ..

stagename:
  image: ..
```
pur parsing yaml then produce a map[string]Stage, which isn't predictible (by intention) when iterating
*/
// Pipeline define the various steps to be executed to produce artifact
type Pipeline map[string]Stage

// UnmarshalYAML implements yaml.v2 Unmarshaler interface to set Stage.Name reflecting map's key
func (p *Pipeline) UnmarshalYAML(unmarshal func(interface{}) error) error {
	ps := map[string]Stage{}
	if err := unmarshal(&ps); err != nil {
		return err
	}
	for k,s := range ps {
		s.Name = k
	}
	*p = ps
	return nil
}


// Stage defines a  set of commands we run in a docker container
type Stage struct {
	StageSpec
	Order    int // injected during yml parsing
	Name	string
} 

type StageSpec struct {
	Image    string
	Env		 map[string]string
	Commands []string
}


var i = 0
// UnmarshalYAML implements yaml.v2 Unmarshaler interface to inject an Order attribute
// while the docker-pipeline yaml file is parsed, as go map isn't ordered (by design).
func (s *Stage) UnmarshalYAML(unmarshal func(interface{}) error) error {
	st := StageSpec{}
	if err := unmarshal(&st); err != nil {
		return err
	}	
	s.Order = i
	s.Image = st.Image
	s.Commands = st.Commands
	s.Env = st.Env
	i=i+1
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
	
	ctx := context.Background()

	env := []string{}
	for k,v := range s.Env {
		env = append(env, k+"="+v)
	}

	// create the container as defined by pipeline's stage
	spec := container.Config{Image: s.Image,
		Tty:          true,
		AttachStdout: true,
		AttachStderr: true,
		WorkingDir:   "/work",
		Cmd:          []string{"/tmp/script.sh"},
		Env:				env,
	}

	c, err := docker.ContainerCreate(ctx, &spec, nil, nil, "stage_"+s.Name)
	if err != nil {
		panic(err)
	}
	defer docker.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{RemoveVolumes: true})

	fmt.Printf(" run stage in %s\n", c.ID)

	// create tar with a single script.sh file containing commands to run
	buf := createCommandsTar(s)
	err = docker.CopyToContainer(ctx, c.ID, "/tmp", buf, types.CopyToContainerOptions{})
	if err != nil {
		panic(err)
	}

	// attach stdour
	resp, err := docker.ContainerAttach(ctx, c.ID, types.ContainerAttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		panic(err)
	}

	receive := make(chan error, 1)
	go func() {
		_, err = io.Copy(os.Stdout, resp.Reader)
		receive <- err
	}()


	// run container
	err = docker.ContainerStart(ctx, c.ID, types.ContainerStartOptions{})
	if err != nil {
		panic(err)
	}
	defer docker.ContainerStop(ctx, c.ID, nil)

	// wait until container has stopped
	filter := filters.NewArgs()
	filter.Add("id", c.ID)

	for {
		time.Sleep(time.Second)
		json, err := docker.ContainerInspect(ctx, c.ID)
		if err != nil {
			panic(err)
		}
		if !json.State.Running {
			if json.State.ExitCode != 0 {
				fmt.Println("[FAILURE]")
				panic(nil) // TODO return error
			}
			break
		}
	}

	/*
		// in parallel, pipe it's log to stdout
		docker.ContainerLogs(ctx, c.ID, types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true})
	*/

	return nil

}

func createCommandsTar(s Stage) *bytes.Buffer {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	bytes := createScript(s)
	header := new(tar.Header)
	header.Name = "script.sh"
	header.Size = int64(len(bytes))
	header.Mode = 0700
	if err := tw.WriteHeader(header); err != nil {
		panic(err)
	}
	if _, err := tw.Write(bytes); err != nil {
		panic(err)
	}
	return buf
}

func createScript(s Stage) []byte {
	var b bytes.Buffer
	b.WriteString("#! /bin/sh\n")
	b.WriteString("set -e\n")
	b.WriteString("set -x\n")

	for _, cmd := range s.Commands {
		b.WriteString(cmd)
		b.WriteString("\n")
	}
	return b.Bytes()
}
