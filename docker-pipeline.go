package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/docker/engine-api/client"
)

func main() {

	var f = flag.String("file", "docker-pipeline.yml", "docker-pipeline yaml file")
	flag.Parse()

	source, err := ioutil.ReadFile(*f)
	if err != nil {
		log.Fatal(err)
	}

	pipeline, err := Parse(source)
	if err != nil {
		log.Fatal(err)
	}

	docker, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err)
	}

	for _, s := range pipeline.Stages() {
		err = runStage(docker, s)
		if err != nil {
			log.Fatal(err)
		}
	}

}

func runStage(docker *client.Client, s Stage) error {
	fmt.Printf("-----------------------------------------\n")
	fmt.Printf(" Stage: %s\n", s.Name)
	fmt.Printf("-----------------------------------------\n")

	return s.Exec.Run(docker, s)
}
