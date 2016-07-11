package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/docker/engine-api/client"
)

func main() {

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	var f = flag.String("file", "docker-pipeline.yml", "docker-pipeline yaml file")
	var project = flag.String("project", filepath.Base(pwd), "name of current project")
	flag.Parse()

	source, err := ioutil.ReadFile(*f)
	if err != nil {
		log.Fatal(err)
	}

	stages, err := Parse(source)
	if err != nil {
		log.Fatal(err)
	}

	pipeline := Pipeline{stages, *project}

	docker, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err)
	}

	action := flag.Arg(0)
	if action == "" {
		action = "run"
	}

	switch action {
	case "run":
		for _, s := range pipeline.Stages() {
			err = runStage(docker, &s, &pipeline)
			if err != nil {
				log.Fatal(err)
			}
		}
	default:
		log.Fatal("Unsupported action " + action)
	}

}

func runStage(docker *client.Client, s *Stage, p *Pipeline) error {
	fmt.Printf("-----------------------------------------\n")
	fmt.Printf(" Stage: %s\n", s.Name)
	fmt.Printf("-----------------------------------------\n")

	return s.Exec.Run(docker, p, s)
}
