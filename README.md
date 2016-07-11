# docker-pipeline

This project is a build orchestration tool for continuous delivery and automation, relying on Docker for execution steps.

Goal of this project is to let end-user model his build pipeline as a set of commands / scripts to run inside docker containers. 
The pipeline is defined as a set of named `Stages`, each of them running a Docker container where configured commands get ran. Optionaly
additional containers can be linked to this build container to offer additional services required by the command (typically: a test 
database, or Selenium browser for functional tests)

This project do define the `docker-pipeline.yml` file format and provide a runner for local usage. This let end-user test his pipeline
locally before committing to project repository. Alternate runners can be created for CI/CD servers, typically using 
[Jenkins' docker-pipeline-plugin](https://github.com/Dockins/docker-pipeline-yml-plugin).

# docker-pipeline.yml reference

`Pipeline` is configured using a yaml file. This file do define a set of named `Stages` a base build blocks. docker-pipeline will run
Stages in sequence following definition ordering.

## Run commands inside a Docker container

This generic stage let you define a docker container to run arbitrary commands. Stage will run commands in sequence
and fail if any of them do return non 0 status. The current diretory is bind mounted inside container as `/work` so
you can access source code stored side by side with your `docker-pipeline.yml` file.

```yaml
build:
    image: ubuntu
    commands:
    -   echo Hello World!
    -   ls -al
```

Container is removed after commands completion, so if you want some files to be cached from build to build, declare 
matching folder as a cached one. A volume is created for those cached paths, and re-used on next run. You can 
typically use this to cache dependency resolution folder for your build tools. 

```yaml
build:
    image: maven:3.3.3-jdk-8
    commands:
    -   mvn package
    cached:
    -   /root/.m2
```

### Environment

You can pass environment variables to the container(s). Can be static values set in `docker-pipeline.yml`, or can be 
infered from your local environment if you prefix them with `$`. 

```yaml
build:
    image: ubuntu
    env:
        FOO: "polka"
        BAR: "$BAR"
    commands:
    -   echo $FOO
    -   echo $BAR
```

```
$ BAR=hello docker-pipeline --file test.yml  
-----------------------------------------
 Stage: build
-----------------------------------------
 run stage in 1680aca2d78c9c5bf031f9e479d53718759608afb10f7e02416f020f98750070
+ echo polka
polka
+ echo hello
hello
```

### Working Directory

You can define the working directory in container using workdir attribute.

```yaml
build:
    image: ubuntu
    workdir: /tmp
    commands:
    -   pwd
```

`workdir` can also be used to let you access a folder in the local working directory. Just prefix workdir
attribute with some path and a `:` separator, it will be used to bind mount the declared path in container. 
If you store `docker-pipeline.yml` file with your project source code in SCM, you'll then be able to access
it to run project build tools inside containers. A relative paths is assumed if you don't explicitely use
`./` notation for sub-folders in your working copy.

```yaml
build:
    image: ubuntu
    workdir: .:/work
    commands:
    -   ls -al
```

### Stash

Using stash is the recommended way to share artifacts between pipeline stages. 
`stash` attribute of a stage define a name and path for artifact to extract from the container on succesful completion. This named
artifact can then be reused by another stage using `unstash` passing stashed artifact name and path inside container to store it 
before the command is ran. If path for a stash isn't absolute, it is considered relative to the working directory    

```yaml
compile:
    image: maven:3.3.3-jdk-8
    workdir: .:/work
    commands:
    -   mvn package
    stash:
        bin:target/app.war

test:
    image:tomcat
    unstash:
        bin:/opt/webapps/app.war
    commands:
    -   ...    
```

# TODO

## Compose multiple containers
You can also declare additional containers to build a complex environment, using a docker-compose like approach
```yaml
build:
    image: maven:3.3.3-jdk-8
    commands:
    -   mvn test
    compose:
        db:
            image: mysql
        selenium:
            image: selenium/standalone-firefox
```

such containers will be linked together and all share a volume (current working directory)


## Build and push Docker images
```yaml
image:
    build:
        dockerfile: Dockerfile.production
        context: . 
        tag: acme/myapp:latest
```