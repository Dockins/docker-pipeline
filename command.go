package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/filters"
	"golang.org/x/net/context"
)

type Command struct {
	Image    string
	Env      map[string]string
	Commands []string
	Cached   []string
	Shell    string
}

func (cmd Command) Run(docker *client.Client, s Stage) error {
	ctx := context.Background()

	env := []string{}
	for k, v := range cmd.Env {
		if v[0] == '$' {
			v = os.Getenv(v[1:])
		}
		env = append(env, k+"="+v)
	}

	// create the container as defined by pipeline's stage
	containerConfig := container.Config{
		Image:        cmd.Image,
		Tty:          true,
		AttachStdout: true,
		AttachStderr: true,
		WorkingDir:   "/work",
		Cmd:          []string{"/tmp/script.sh"},
		Env:          env,
	}

	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	binds := []string{pwd + ":/work"}

	// for each path in cache, retrieve or create a docker volume
	for _, v := range cmd.Cached {
		// TODO generate a unique volume ID based on path
		vid := "polka_foobar"
		filter := filters.NewArgs()
		filter.Add("name", vid)
		volumes, err := docker.VolumeList(ctx, filter)
		if err != nil {
			panic(err)
		}
		if len(volumes.Volumes) == 0 {
			_, err := docker.VolumeCreate(ctx, types.VolumeCreateRequest{
				Name: vid,
			})
			if err != nil {
				panic(err)
			}
		}
		fmt.Println("mount volume " + vid + " to cache " + v)
		binds = append(binds, vid+":"+v)
	}

	hostConfig := container.HostConfig{
		Binds: binds,
	}

	c, err := docker.ContainerCreate(ctx, &containerConfig, &hostConfig, nil, "stage_"+s.Name)
	if err != nil {
		panic(err)
	}
	defer docker.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{RemoveVolumes: true})

	fmt.Printf(" run stage in %s\n", c.ID)

	// create tar with a single script.sh file containing commands to run
	buf := cmd.createCommandsTar()
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

func (cmd *Command) createCommandsTar() *bytes.Buffer {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	bytes := cmd.createScript()
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

func (cmd *Command) createScript() []byte {
	var b bytes.Buffer
	sh := "bash"
	if cmd.Shell != "" {
		sh = cmd.Shell
	}
	b.WriteString("#! /bin/" + sh + "\n")
	b.WriteString("set -e\n")
	b.WriteString("set -x\n")

	for _, cmd := range cmd.Commands {
		b.WriteString(cmd)
		b.WriteString("\n")
	}
	return b.Bytes()
}
