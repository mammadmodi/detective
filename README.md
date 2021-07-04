# Detective

[![Workflow Status](https://github.com/mammadmodi/detective/workflows/Test/badge.svg)](https://github.com/mammadmodi/detective/actions)
[![codecov](https://codecov.io/gh/mammadmodi/detective/branch/main/graph/badge.svg)](https://codecov.io/gh/mammadmodi/detective)
[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/mammadmodi/detective?filename=go.mod)](https://github.com/mammadmodi/detective)
[![Docker Image Size (tag)](https://img.shields.io/docker/image-size/mammadmodi/detective/main?style=flat-square&logo=docker)](https://github.com/mammadmodi/detective/blob/main/build/Dockerfile)
[![Docker Pulls](https://img.shields.io/docker/pulls/mammadmodi/detective?style=flat-square&logo=docker)](https://hub.docker.com/r/mammadmodi/detective/tags?page=1&ordering=last_updated)

The web application for analyzing web pages written with golang.

~~~
 ______   _______ _________ _______  _______ __________________          _______ 
(  __  \ (  ____ \\__   __/(  ____ \(  ____ \\__   __/\__   __/|\     /|(  ____ \
| (  \  )| (    \/   ) (   | (    \/| (    \/   ) (      ) (   | )   ( || (    \/
| |   ) || (__       | |   | (__    | |         | |      | |   | |   | || (__    
| |   | ||  __)      | |   |  __)   | |         | |      | |   ( (   ) )|  __)   
| |   ) || (         | |   | (      | |         | |      | |    \ \_/ / | (      
| (__/  )| (____/\   | |   | (____/\| (____/\   | |   ___) (___  \   /  | (____/\
(______/ (_______/   )_(   (_______/(_______/   )_(   \_______/   \_/   (_______/
~~~

## What is detective?

Detective is a web application by which you can get useful information about an url. First it will perform an GET
request to entry url and then retrieves the following information:

1. Version of HTML
2. Page Title
3. Count of Heading tags (h1-h6)
4. Count of links (internal and external)
5. Count of inaccessible links
6. Existence of Login Form

---
**NOTE**

It's a sample golang application, so I tried to make its folders and package structured. You can find the full
declaration of this folder structure at [this repo](https://github.com/golang-standards/project-layout). It's not
recommended using this folder structure for small projects.

---

## How to set up?

### Docker

Use `docker run --rm -p 8000:8000 mammadmodi/detective:main` command to create a container in the foreground.Just be
aware that the `mammadmodi/detective:main` docker image has been built only for **linux/amd64** Arch so if you use
another architecture then try to use one of the following ways.

### Docker Compose

Clone the project in your local and run the app easily by `make up` command in your desktop to set up detective
application. It first builds a docker image in your local and then runs an instance of that in the background using
**docker compose**.

Don't forget to run `make down` when you finished your work with this app.

### Go Compiler

[Install Go](https://golang.org/doc/install) in your local and use the bellow command to set up application server:

`go run ./cmd/server/main.go`

## How To Use?

Anyway, when you set up the application, it will be started on port 8000 by default, and you can use
its [Form](http://127.0.0.1:8000/analyze-url.html) to analyze your web pages.

### App Configuration

You can configure the application by setting environment variables on your os. The list of available configurations has
shown below:

| **Variable Name**              | **Type** | **Default** |                **Description**                  |
| ------------------------------ | -------- | ----------- | ----------------------------------------------- |
| `DETECTIVE_ADDR`        | ***string***  | ":8000" | The address of http server with its port |
| `DETECTIVE_HTTP_TIMEOUT` | ***string*** | "30s" | Timeout for performing http requests |
| `DETECTIVE_LOGGER_ENABLED` | ***boolean*** | true | Feature flag for logger|
| `DETECTIVE_LOGGER_LEVEL` | ***string*** | "info" | Level of logger in string format(debug,info,warn,...)|
| `DETECTIVE_LOGGER_PRETTY` | ***boolean*** | true | If set to false logs will be structured in json objects|
| `DETECTIVE_LOGGER_FILE_REDIRECT_ENABLED`  | ***false***  | false | Feature flag for storing logs to files|
| `DETECTIVE_LOGGER_FILE_REDIRECT_PATH` | ***string***  | "/var/log" | Directory of log file storage|
| `DETECTIVE_LOGGER_FILE_REDIRECT_PREFIX` | ***string*** | "detective" | Prefix for log files |
