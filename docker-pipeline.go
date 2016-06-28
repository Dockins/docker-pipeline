package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/docker/engine-api/client"
)

func main() {

	var f = flag.String("file", "docker-pipeline.yml", "docker-pipeline yaml file")

	source, err := ioutil.ReadFile(*f)
	if err != nil {
		panic(err)
	}

	pipeline, err := Parse(source)
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
