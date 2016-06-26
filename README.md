# docker-pipeline

This project is a build orchestration tool for continuous delivery and automation, relying on Docker for execution steps.

Goal of this project is to let end-user model his build pipeline as a set of commands / scripts to run inside docker containers. 
The pipeline is defined as a set of named `Stages`, each of them running a Docker container where configured commands get ran. Optionaly
additional containers can be linked to this build container to offer additional services required by the command (typically: a test 
database, or Selenium browser for functional tests)

This project do define the `docker-pipeline.yml` file format and provide a runner for local usage. This let end-user test his pipeline
locally before committing to project repository. Alternate runners can be created for CI/CD servers, typically using 
[Jenkins' docker-pipeline-plugin](https://github.com/Dockins/docker-pipeline-plugin)

# docker-pipeline.yml reference

`Pipeline` is configured using a yaml file. This file do define a set of named `Stages` a base build blocks. A Stage do declare:
- a Docker image ID for the container to run the commands
- a set of commands to run in this container.

pipeline will run Stages in sequence following definition ordering
