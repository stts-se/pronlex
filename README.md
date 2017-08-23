# pronlex
pronlex is a pronunciation lexicon database with a server behind an http API.

[![GoDoc](https://godoc.org/github.com/stts-se/pronlex?status.svg)](https://godoc.org/github.com/stts-se/pronlex)


## Docker installation

TODO/WORK IN PROGRESS


`$ cd <DOCKER DIR>`   

`$ docker build --no-cache https://raw.githubusercontent.com/stts-se/pronlex/master/Dockerfile \
-t sttsse/lexserver`

OR

`$ docker build --no-cache $GOPATH/src/github.com/stts-se/pronlex \
-t sttsse/lexserver`

`$ docker run -v <DOCKERDIR>/lexserver_files:/go/lexserver_files -p 8787:8787 -it sttsse/lexserver sh import_lex`


## Lexicon server and setup

* **Standalone setup**    
Readme: install/standalone/README.md   
Git folder: [install/standalone/](https://github.com/stts-se/pronlex/blob/master/install/standalone)

* **Setup for developers**    
Readme: install/developer/README.md   
Git folder: [install/developer/](https://github.com/stts-se/pronlex/blob/master/install/developer)



## Regexp db search performance

Regular expression search using a Go's regular expressions through the Sqlite3 driver is very slow. Either we should change databases, or find a better way to do regexp search in Sqlite3 from Go.


