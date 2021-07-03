# Detective

[![Workflow Status](https://github.com/mammadmodi/detective/workflows/Test/badge.svg)](https://github.com/mammadmodi/detective/actions)
[![codecov](https://codecov.io/gh/mammadmodi/detective/branch/main/graph/badge.svg)](https://codecov.io/gh/mammadmodi/detective)
[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/mammadmodi/detective?filename=go.mod)](https://github.com/mammadmodi/detective)

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

Easily you can use `make up` command to set up detective application. First it will build the docker image of
application, and then it runs an instance of that in background using docker compose. The application will be start on
port 8000 by default, and you can use [it's form](http://127.0.0.1:8000/analyze-url) to analyze your web pages.