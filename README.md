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

`Pipeline` is configured using a yaml file. This file do define a set of named `Stages` a base build blocks. docker-pipeline will run
Stages in sequence following definition ordering.

## Run commands inside a Docker container

This generic stage let you define a docker container to run arbitrary commands. Stage will run commands in sequence
and fail if any of them do return non 0 status. The current diretory is bind mounted inside container as `/work` so
you can access source code stored side by side with your `docker-pipeline.yml` file.

```
build:
    image: ubuntu
    commands:
    -   echo Hello World!
    -   ls -al
```

Container is removed after commands completion, so if you want some files to be cached from build to build, declare 
matching folder as a cached one. A volume is created for those cached paths, and re-used on next run. You can 
typically use this to cache dependency resolution folder for your build tools. 

```
build:
    image: maven:3.3.3-jdk-8
    commands:
    -   mvn package
    cached:
    -   /root/.m2
```

You can pass environment variables to the container(s). Can be static values set in `docker-pipeline.yml`, or can be 
infered from your local environment if you prefix them with `$`. 

```
build:
    image: ubuntu
    commands:
    -   echo $FOO
    -   echo $BAR
    end:
    -   FOO=polka
    -   BAR=$BAR
```